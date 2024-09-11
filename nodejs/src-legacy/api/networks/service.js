export default class NetworkService {
    #networkName;

    constructor(networkName) {
        this.#networkName = networkName;
    }

    getUniswapQuoterAddress() {
        if (this.#networkName === 'mainnet') {
            return '0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6';
        } else if (this.#networkName === 'base') {
            return '';
        }
    }
}
