import pLimit from 'p-limit';
import { writeFile } from '../../common/helpers.js';
import { DATA_DIR } from '../../common/consts.js';

export default class TokenPriceAggregationService {
    #uniswapService;
    #interval;
    #tokenPairs;
    #pairPrices;
    #shouldWriteFileAction;
    #ethereumPairs = [
        'USDC_WETH', 'WBTC_WETH', 'USDC_WETH', 'WETH_USDT', 'WBTC_USDC', 'WBTC_WETH', 'WETH_USDT', 'DAI_USDC', 'FRAX_USDC', 'DAI_WETH', 'LINK_WETH', 'PEPE_WETH', 'WBTC_USDT', 'HKDM_USDM', 'USDC_DYAD', 'MKR_WETH', 'UNI_WETH', 'WHITE_WETH', 'USDC_USDT', 'PEPE_WETH', 'WETH_mETH', 'WETH_ETHM', 'MKR_WETH', 'wstETH_WETH', 'USDM_USDT', 'MNT_WETH', 'TURBO_WETH', 'USDe_USDT', 'WETH_weETH', 'DAI_WETH', 'WETH_weETH', 'DAI_USDC', 'SHIB_WETH', 'DOG_WETH', 'rsETH_WETH', 'APE_WETH', 'WETH_ENS', 'PANDORA_WETH', 'WETH_SOL', 'DAI_FRAX', 'LDO_WETH', 'MATIC_WETH', 'AAVE_WETH', 'WETH_ONDO', 'USDC_USDT', 'WETH_LOOKS', 'WETH_USDT', 'PEOPLE_WETH', 'SHFL_USDC', 'USDC_WETH', 'FTM_WETH', 'RNDR_WETH', 'KOIN_USDC', 'RCH_WETH', 'FET_WETH', 'SHIB_WETH', 'WBTC_LBTC', 'OX_OX', 'PORK_WETH', 'PRIME_WETH', 'HEX_USDC'
    ];

    /**
     * @param uniswapService {UniswapService}
     * @param interval {number}
     */
    constructor(uniswapService, interval) {
        this.#uniswapService = uniswapService;
        this.#interval = interval;
        this.#tokenPairs = this.#getSuitablePairs();
        this.#pairPrices = {};
        this.#shouldWriteFileAction = false;
    }

    #getSuitablePairs() {
        const pairsInfo = Object.values(this.#uniswapService.getTokenInfo());
        const results = [];

        // determine the suitable pairs (maybe by symbol, or by some other criteria)
        for (const pair of pairsInfo) {
            const pairSymbol = `${pair.token0.symbol}_${pair.token1.symbol}`;

            if (this.#ethereumPairs.includes(pairSymbol)) {
                results.push(pair);
            }
        }

        return results;
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
                    const price = await this.#uniswapService.getPrice(pair.pairAddress, 1, 'baseToQuote');
                    const pairSymbol = `${pair.token0.symbol}_${pair.token1.symbol}`;
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

    async stop() {
        // do something
    }
}