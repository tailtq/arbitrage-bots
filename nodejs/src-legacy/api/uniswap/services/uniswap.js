import * as ethers from 'ethers';
import fs from 'fs';
import pLimit from 'p-limit';
import Quoter from '@uniswap/v3-periphery/artifacts/contracts/lens/Quoter.sol/Quoter.json' assert { type: 'json' };
import { LOGS_DIR } from '../../common/consts.js';
import { writeFile } from '../../common/helpers.js';

const QuoterABI = Quoter.abi;

class UniswapService {
    #provider;
    #quoterAddress;
    #quoterContract;
    #tokenInfo;
    #logger;
    #tokenInfoCacheFile = 'tokenInfo.json';

    static UNISWAP_PAIR_ABI = [
        'function token0() external view returns (address)',
        'function token1() external view returns (address)',
        'function fee() external view returns (uint24)',
    ];

    static TOKEN_ABI = [
        'function symbol() external view returns (string)',
        'function name() external view returns (string)',
        'function decimals() external view returns (uint)',
    ];

    constructor() {
        this.#provider = new ethers.JsonRpcProvider(process.env.NETWORK_CONNECTION_URL);
        this.#provider._getConnection().timeout = 30000;
        this.#quoterAddress = process.env.QUOTER_ADDRESS;
        this.#tokenInfo = this.#loadTokenInfoLocal();
        this.#quoterContract = new ethers.Contract(this.#quoterAddress, QuoterABI, this.#provider);
        this.#logger = fs.createWriteStream(LOGS_DIR + '/arbitrage-logs.jsonl', { flags: 'a' });
    }

    getTokenInfo() {
        return this.#tokenInfo;
    }

    async loadTokens(pairAddresses) {
        const limit = pLimit(1);
        const { TOKEN_ABI, UNISWAP_PAIR_ABI } = this.constructor;

        for (const pairAddress of new Set(pairAddresses)) {
            console.log(`Get token list ${pairAddress}`);
            if (this.#tokenInfo[pairAddress]) continue;

            try {
                console.time(`Get token list ${pairAddress}`);
                const poolContract = new ethers.Contract(pairAddress, UNISWAP_PAIR_ABI, this.#provider);
                const [token0Address, token1Address, fee] = await Promise.all([
                    poolContract.token0(),
                    poolContract.token1(),
                    poolContract.fee()
                ]);
                const tokenInfo = {
                    pairAddress,
                    fee,
                    token0: { address: token0Address },
                    token1: { address: token1Address },
                };

                for (const tokenAddress of [token0Address, token1Address]) {
                    const tokenContract = new ethers.Contract(tokenAddress, TOKEN_ABI, this.#provider);
                    const [tokenSymbol, tokenName, tokenDecimals] = await Promise.all([
                        tokenContract.symbol(),
                        tokenContract.name(),
                        tokenContract.decimals(),
                    ]);
                    const tokenData = {
                        address: tokenAddress,
                        symbol: tokenSymbol,
                        name: tokenName,
                        decimals: tokenDecimals,
                    };

                    if (tokenInfo.token0.address === tokenAddress) tokenInfo.token0 = tokenData;
                    if (tokenInfo.token1.address === tokenAddress) tokenInfo.token1 = tokenData;
                }

                await writeFile(this.#tokenInfoCacheFile, this.#tokenInfo);
                console.timeEnd(`Get token list ${pairAddress}`);
            } catch (e) {
                console.error(e);
            }
        }

        console.log('Token info loaded');
        await writeFile(this.#tokenInfoCacheFile, this.#tokenInfo);

        return this.#tokenInfo;
    }

    #loadTokenInfoLocal() {
        if (fs.existsSync(this.#tokenInfoCacheFile)) {
            const data = fs.readFileSync(this.#tokenInfoCacheFile);
            return JSON.parse(data);
        }
        return {};
    }

    async getBatchDepthOpportunity(surfaceResults) {
        const limit = pLimit(2);
        const promises = surfaceResults.map(surfaceResult => limit(async () => {
            const [resultForward, resultBackward] = await Promise.all([
                this.getDepthOpportunityForward(surfaceResult),
                this.getDepthOpportunityBackward(surfaceResult),
            ]);
            const depthResult = {
                forward: resultForward,
                backward: resultBackward,
            };
            this.logArbOpportunity({ surfaceResult, depthResult });
            const key = `${surfaceResult.swap1}_${surfaceResult.swap2}_${surfaceResult.swap3}`;

            return {
                [`${key}`]: depthResult,
            };
        }));
        const results = await Promise.all(promises);
        const resultObj = results.reduce((acc, result) => ({ ...acc, ...result }), {});

        return resultObj;
    }

    async getDepthOpportunityForward(surfaceResult) {
        const contract1 = surfaceResult.contract1;
        const contract2 = surfaceResult.contract2;
        const contract3 = surfaceResult.contract3;
        const pair1ContractAddress = surfaceResult.contract1Address;
        const pair2ContractAddress = surfaceResult.contract2Address;
        const pair3ContractAddress = surfaceResult.contract3Address;
        const directionTrade1 = surfaceResult.directionTrade1;
        const directionTrade2 = surfaceResult.directionTrade2;
        const directionTrade3 = surfaceResult.directionTrade3;
        const allContracts = `${contract1}_${contract2}_${contract3}`;

        console.log('Checking trade 1 acquired coin...');
        const acquiredCoinT1 = await this.getPrice(pair1ContractAddress, surfaceResult.startingAmount, directionTrade1);
        console.log(allContracts, 'acquiredCoinT1', acquiredCoinT1, 'startingAmount', surfaceResult.startingAmount);

        // console.log('Checking trade 2 acquired coin...');
        if (acquiredCoinT1 === 0) return;
        const acquiredCoinT2 = await this.getPrice(pair2ContractAddress, acquiredCoinT1, directionTrade2);
        console.log(allContracts, 'acquiredCoinT2', acquiredCoinT2, 'startingAmount', acquiredCoinT1);

        // console.log('Checking trade 3 acquired coin...');
        if (acquiredCoinT2 === 0) return;
        const acquiredCoinT3 = await this.getPrice(pair3ContractAddress, acquiredCoinT2, directionTrade3);
        console.log(
            allContracts,
            'acquiredCoinT3',
            acquiredCoinT3,
            'startingAmount',
            acquiredCoinT2,
            'direction',
            directionTrade1,
            this.#calculateArb(surfaceResult.startingAmount, acquiredCoinT3, surfaceResult)
        );

        // Calculate and show result
        return {
            contract1,
            contract2,
            contract3,
            directionTrade1,
            directionTrade2,
            directionTrade3,
            acquiredCoinT1: parseFloat(acquiredCoinT1),
            acquiredCoinT2: parseFloat(acquiredCoinT2),
            acquiredCoinT3: parseFloat(acquiredCoinT3),
            ...this.#calculateArb(surfaceResult.startingAmount, acquiredCoinT3, surfaceResult)
        };
    }

    async getDepthOpportunityBackward(surfaceResult) {
        const contract1 = surfaceResult.contract3;
        const contract2 = surfaceResult.contract2;
        const contract3 = surfaceResult.contract1;
        const pair1ContractAddress = surfaceResult.contract3Address;
        const pair2ContractAddress = surfaceResult.contract2Address;
        const pair3ContractAddress = surfaceResult.contract1Address;
        const directionTrade1 = this.#revertDirection(surfaceResult.directionTrade3);
        const directionTrade2 = this.#revertDirection(surfaceResult.directionTrade2);
        const directionTrade3 = this.#revertDirection(surfaceResult.directionTrade1);

        console.log('Checking trade 1 acquired coin...');
        const acquiredCoinT1 = await this.getPrice(pair1ContractAddress, surfaceResult.startingAmount, directionTrade1);

        // console.log('Checking trade 2 acquired coin...');
        if (acquiredCoinT1 === 0) return;
        const acquiredCoinT2 = await this.getPrice(pair2ContractAddress, acquiredCoinT1, directionTrade2);

        // console.log('Checking trade 3 acquired coin...');
        if (acquiredCoinT2 === 0) return;
        const acquiredCoinT3 = await this.getPrice(pair3ContractAddress, acquiredCoinT2, directionTrade3);

        // Calculate and show result
        return {
            contract1,
            contract2,
            contract3,
            directionTrade1,
            directionTrade2,
            directionTrade3,
            acquiredCoinT1: parseFloat(acquiredCoinT1),
            acquiredCoinT2: parseFloat(acquiredCoinT2),
            acquiredCoinT3: parseFloat(acquiredCoinT3),
            ...this.#calculateArb(surfaceResult.startingAmount, acquiredCoinT3, surfaceResult)
        };
    }

    #revertDirection(direction) {
        if (direction === 'baseToQuote') return 'quoteToBase';
        if (direction === 'quoteToBase') return 'baseToQuote';
    }

    // GET PRICE /////////////////////////////////////////////////
    async getPrice(address, amountIn, tradeDirection, verbose = true) {
        if (!this.#tokenInfo[address]) {
            // console.error('Token info not found');
            return 0;
        }

        const { fee, token0, token1 } = this.#tokenInfo[address];
        let inputTokenA;
        let inputDecimalA;
        let inputSymbolA;
        let inputTokenB;
        let inputDecimalB;
        let inputSymbolB;

        if (tradeDirection === 'baseToQuote') {
            inputTokenA = token0.address;
            inputSymbolA = token0.symbol;
            inputDecimalA = token0.decimals;
            inputTokenB = token1.address;
            inputSymbolB = token1.symbol;
            inputDecimalB = token1.decimals;
        } else if (tradeDirection === 'quoteToBase') {
            inputTokenA = token1.address;
            inputSymbolA = token1.symbol;
            inputDecimalA = token1.decimals;
            inputTokenB = token0.address;
            inputSymbolB = token0.symbol;
            inputDecimalB = token0.decimals;
        }

        // reformat amountIn
        const amountInParsed = BigInt(parseInt(amountIn * 10 ** inputDecimalA));
        // console.time(`quoteExactInputSingle_${inputSymbolA}${inputSymbolB}`);

        try {
            let quotedAmountOut = await this.#quoterContract.quoteExactInputSingle.staticCall(
                inputTokenA,
                inputTokenB,
                fee,
                amountInParsed,
                0
            );
            quotedAmountOut = ethers.formatUnits(quotedAmountOut, parseInt(inputDecimalB));

            return quotedAmountOut;
        } catch (err) {
            // if (verbose) console.error('Quoter error:', err);
            console.log('Error with params', inputSymbolA, inputSymbolB, fee, amountInParsed, 0);
            return 0;
        } finally {
            // console.timeEnd(`quoteExactInputSingle_${inputSymbolA}${inputSymbolB}`);
        }
    }

    #calculateArb(amountIn, outputOut) {
        const profitLoss = outputOut - amountIn;
        const profitLossPerc = (profitLoss / amountIn) * 100;

        return {
            profitLoss,
            profitLossPerc,
        };
    }

    logArbOpportunity(data) {
        this.#logger.write(JSON.stringify(data) + '\n');
    }
}

export default UniswapService;
