const ethers = require('ethers');
const QuoterABI = require('@uniswap/v3-periphery/artifacts/contracts/lens/Quoter.sol/Quoter.json').abi;

class UniswapService {
    #provider;
    #quoterAddress;
    #quoterContract;
    #tokenInfo;

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
        this.#tokenInfo = {};
        this.#quoterContract = new ethers.Contract(this.#quoterAddress, QuoterABI, this.#provider);
    }

    async loadTokens(pairAddresses) {
        const promises = pairAddresses.map(async pairAddress => {
            const uniswapPairABI = this.constructor.UNISWAP_PAIR_ABI;
            const poolContract = new ethers.Contract(pairAddress, uniswapPairABI, this.#provider);
            const [token0Address, token1Address, fee] = await Promise.all([
                poolContract.token0(),
                poolContract.token1(),
                poolContract.fee()
            ]);
            const addressArray = [token0Address, token1Address];
            const tokenInfo = {
                pairAddress,
                fee,
                token0: {
                    address: token0Address,
                },
                token1: {
                    address: token1Address,
                },
            };

            for (let j = 0; j < addressArray.length; j++) {
                const tokenAddress = addressArray[j];
                const tokenABI = this.constructor.TOKEN_ABI;

                const tokenContract = new ethers.Contract(tokenAddress, tokenABI, this.#provider);
                const [tokenSymbol, tokenName, tokenDecimals] = await Promise.all([
                    tokenContract.symbol(),
                    tokenContract.name(),
                    tokenContract.decimals(),
                ]);
                const tokenData = {
                    id: `token${j}`,
                    address: tokenAddress,
                    symbol: tokenSymbol,
                    name: tokenName,
                    decimals: tokenDecimals,
                };

                if (tokenInfo.token0.address === tokenAddress) tokenInfo.token0 = tokenData;
                if (tokenInfo.token1.address === tokenAddress) tokenInfo.token1 = tokenData;
            }

            return this.#tokenInfo[pairAddress] = tokenInfo;
        });

        return Promise.all(promises);
    }

    async getDepthOpportunity(surfaceResult, amountIn) {
        const pair1ContractAddress = surfaceResult.contract1Address;
        const pair2ContractAddress = surfaceResult.contract2Address;
        const pair3ContractAddress = surfaceResult.contract3Address;
        const directionTrade1 = surfaceResult.directionTrade1;
        const directionTrade2 = surfaceResult.directionTrade2;
        const directionTrade3 = surfaceResult.directionTrade3;

        console.log('Checking trade 1 acquired coin...');
        const acquiredCoinT1 = await this.#getPrice(pair1ContractAddress, 0.1, directionTrade1);

        // console.log('Checking trade 2 acquired coin...');
        if (acquiredCoinT1 === 0) return;
        const acquiredCoinT2 = await this.#getPrice(pair2ContractAddress, acquiredCoinT1, directionTrade2);

        // console.log('Checking trade 3 acquired coin...');
        if (acquiredCoinT2 === 0) return;
        const acquiredCoinT3 = await this.#getPrice(pair3ContractAddress, acquiredCoinT2, directionTrade3);

        // Calculate and show result
        return this.#calculateArb(amountIn, acquiredCoinT3, surfaceResult);
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
        let amountInParsed = ethers.parseUnits(amountIn.toString(), inputDecimalA);
        let quotedAmountOut = 0;

        try {
            console.log(inputTokenA, inputTokenB, fee, amountInParsed, 0);
            quotedAmountOut = await this.#quoterContract.quoteExactInputSingle.staticCall(
                inputTokenA,
                inputTokenB,
                fee,
                amountInParsed,
                0
            );
            quotedAmountOut = ethers.formatUnits(quotedAmountOut, inputDecimalB);
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

module.exports = UniswapService;
