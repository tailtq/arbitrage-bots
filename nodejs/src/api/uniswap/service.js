import * as ethers from 'ethers';
import fs from 'fs';
import pLimit from 'p-limit';
import Quoter from '@uniswap/v3-periphery/artifacts/contracts/lens/Quoter.sol/Quoter.json' assert { type: 'json' };

const QuoterABI = Quoter.abi;

class UniswapService {
    #provider;
    #quoterAddress;
    #quoterContract;
    #tokenInfo;
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
        this.#provider = new ethers.JsonRpcProvider(`https://mainnet.infura.io/v3/${process.env.INFURA_API_KEY}`);
        this.#quoterAddress = '0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6';
        this.#tokenInfo = this.#loadTokenInfoLocal();
        this.#quoterContract = new ethers.Contract(this.#quoterAddress, QuoterABI, this.#provider);
    }

    async loadTokens(pairAddresses) {
        const limit = pLimit(3);
        const { TOKEN_ABI, UNISWAP_PAIR_ABI } = this.constructor;

        const promises = [... new Set(pairAddresses)].map(async pairAddress => limit(async () =>{
            if (this.#tokenInfo[pairAddress]) return this.#tokenInfo[pairAddress];

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
            console.timeEnd(`Get token list ${pairAddress}`);

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

            return this.#tokenInfo[pairAddress] = tokenInfo;
        }));
        const result = await Promise.all(promises);
        await this.#saveTokenInfoLocal();

        return result;
    }

    async #saveTokenInfoLocal() {
        return new Promise(async (resolve, reject) => {
            const tokenInfo = JSON.stringify(this.#tokenInfo, null, 2);
            fs.writeFile(this.#tokenInfoCacheFile, tokenInfo, () => {
                resolve();
            });
        });
    }

    #loadTokenInfoLocal() {
        if (fs.existsSync(this.#tokenInfoCacheFile)) {
            const data = fs.readFileSync(this.#tokenInfoCacheFile);
            return JSON.parse(data);
        }
        return {};
    }

    async getDepthOpportunity(surfaceResult) {
        const pair1ContractAddress = surfaceResult.contract1Address;
        const pair2ContractAddress = surfaceResult.contract2Address;
        const pair3ContractAddress = surfaceResult.contract3Address;
        const directionTrade1 = surfaceResult.directionTrade1;
        const directionTrade2 = surfaceResult.directionTrade2;
        const directionTrade3 = surfaceResult.directionTrade3;

        console.log('Checking trade 1 acquired coin...');
        const acquiredCoinT1 = await this.#getPrice(pair1ContractAddress, surfaceResult.startingAmount, directionTrade1);

        // console.log('Checking trade 2 acquired coin...');
        if (acquiredCoinT1 === 0) return;
        const acquiredCoinT2 = await this.#getPrice(pair2ContractAddress, acquiredCoinT1, directionTrade2);

        // console.log('Checking trade 3 acquired coin...');
        if (acquiredCoinT2 === 0) return;
        const acquiredCoinT3 = await this.#getPrice(pair3ContractAddress, acquiredCoinT2, directionTrade3);

        // Calculate and show result
        return this.#calculateArb(surfaceResult.startingAmount, acquiredCoinT3, surfaceResult);
    }

    // GET PRICE /////////////////////////////////////////////////
    async #getPrice(address , amountIn, tradeDirection, verbose = true) {
        if (!this.#tokenInfo[address]) {
            console.error('Token info not found');
            return 0;
        }

        console.time('quoteExactInputSingle');
        const { fee, token0, token1 } = this.#tokenInfo[address];
        let inputTokenA;
        let inputDecimalA;
        let inputTokenB;
        let inputDecimalB;

        if (tradeDirection === 'baseToQuote') {
            inputTokenA = token0.address;
            inputDecimalA = token0.decimals;
            inputTokenB = token1.address;
            inputDecimalB = token1.decimals;
        } else if (tradeDirection === 'quoteToBase') {
            inputTokenA = token1.address;
            inputDecimalA = token1.decimals;
            inputTokenB = token0.address;
            inputDecimalB = token0.decimals;
        }

        // reformat amountIn
        const amountInParsed = ethers.parseUnits(amountIn.toString(), parseInt(inputDecimalA));

        try {
            let quotedAmountOut = await this.#quoterContract.quoteExactInputSingle.staticCall(
                inputTokenA,
                inputTokenB,
                fee,
                amountInParsed,
                0
            );
            quotedAmountOut = ethers.formatUnits(quotedAmountOut, parseInt(inputDecimalB));
            console.timeEnd('quoteExactInputSingle');

            return quotedAmountOut
        } catch (err) {
            if (verbose) console.error('Quoter error:', err);
            console.timeEnd('quoteExactInputSingle');
            return 0
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
}

export default UniswapService;
