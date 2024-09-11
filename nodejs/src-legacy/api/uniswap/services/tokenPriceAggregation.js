import pLimit from 'p-limit';
import { writeFile } from '../../common/helpers.js';
import { DATA_DIR } from '../../common/consts.js';

export default class TokenPriceAggregationService {
    #uniswapService;
    #tokenPairs;
    #pairPrices;
    #interval;
    #amountIn;
    #shouldWriteFileAction;
    #ethereumTokenSymbols = [
        'DAI', 'FRAX', 'FEI', 'LINK', 'DAI', 'PEPE', 'DYAD', 'MKR', 'UNI', 'WHITE', 'MNT', 'TURBO', 'SHIB', 'DOG', 'APE', 'ENS', 'PANDORA', 'SOL', 'LDO', 'MATIC', 'AAVE', 'ONDO', 'PEOPLE', 'SHFL', 'FTM', 'RNDR', 'KOIN', 'RCH', 'FET', 'LBTC', 'PORK', 'PRIME', 'HEX'
    ];

    /**
     * @param uniswapService {UniswapService}
     * @param interval {number}
     * @param amountIn {number}
     */
    constructor(uniswapService, interval, amountIn) {
        this.#uniswapService = uniswapService;
        this.#interval = interval;
        this.#amountIn = amountIn;
        this.#tokenPairs = Object.values(this.#uniswapService.getTokenInfo());
        this.#pairPrices = {};
        this.#shouldWriteFileAction = false;
    }

    async startPricePolling() {
        const limit = pLimit(5);
        // Running the write file action in parallel
        this.startWriteFileAction();

        while (true) {
            console.log('Aggregating token prices...');
            const promises = [];

            for (const pair of this.#tokenPairs) {
                promises.push(limit(async () => {
                    const price = await this.#uniswapService.getPrice(pair.pairAddress, this.#amountIn, 'baseToQuote');
                    const pairSymbol = `${pair.token0.symbol}${pair.token1.symbol}`;
                    this.#pairPrices[pairSymbol] = price;
                    this.#shouldWriteFileAction = true;
                    // await writeFile(DATA_DIR + '/tokenPrices.json', this.#pairPrices);
                }));
            }

            await Promise.all(promises);
            console.log('Token prices aggregated');
            await new Promise(resolve => setTimeout(resolve, this.#interval));
        }
    }

    async startWriteFileAction() {
        while (true) {
            if (this.#shouldWriteFileAction) {
                await writeFile(DATA_DIR + '/tokenPrices.json', this.#pairPrices);
                this.#shouldWriteFileAction = false;
            }

            await new Promise(resolve => setTimeout(resolve, 100));
        }
    }

    getPairPrices(tokenPairs) {
        const results = {};

        for (const pairSymbol of tokenPairs) {
            if (!this.#pairPrices[pairSymbol]) {
                console.log(`Pair ${pairSymbol} not found`);
                continue;
            }
            results[pairSymbol] = {
                symbol: pairSymbol,
                token0Price: (this.#amountIn / this.#pairPrices[pairSymbol]).toString(),
                token1Price: this.#pairPrices[pairSymbol],
            };
        }

        return results;
    }

    async stop() {
        // do something
    }
}
