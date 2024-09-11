require('@nomicfoundation/hardhat-toolbox');

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    compilers: [
      { version: '0.5.5', },
      { version: '0.6.6', },
      { version: '0.8.8', },
    ],
  },
  networks: {
    hardhat: {
      forking: {
        url: 'https://bsc-dataseed1.binance.org/',
        blockNumber: 42163104, // block number of the latest block at the moment (without this, it gets error)
      },
    },
    testnet: {
      url: 'https://data-seed-prebsc-1-s1.binance.org:8545/',
      chainId: 97,
      accounts: ['0x92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e'],
    },
    mainnet: {
      url: 'https://bsc-dataseed1.binance.org/',
      chainId: 56,
      accounts: ['0x92db14e403b83dfe3df233f83dfa3a0d7096f21ca9b0d6d6b8d88b2b4ec1564e'],
    },
  },
};
