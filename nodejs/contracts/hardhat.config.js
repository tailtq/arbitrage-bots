require('@nomiclabs/hardhat-waffle');

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    compilers: [
      { version: '0.5.5', },
      { version: '0.6.6', },
      { version: '0.8.8', },
      { version: '0.8.18', },
    ],
  },
  networks: {
    hardhat: {
      forking: {
        url: 'https://bsc-mainnet.infura.io/v3/dd7bf6951267477f8f20d852e0135f84',
        // blockNumber: 42163104, // block number of the latest block at the moment (without this, it gets error)
      },
    },
    testnet: {
      url: 'https://data-seed-prebsc-1-s1.binance.org:8545/',
      chainId: 97,
      accounts: [],
    },
    mainnet: {
      url: 'https://bsc-dataseed1.binance.org/',
      chainId: 56,
      accounts: [],
    },
  },
};
