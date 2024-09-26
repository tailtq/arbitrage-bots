require('@nomiclabs/hardhat-waffle');

const PRIVATE_KEY_TEST = '1bc10d8ef0bdeb8aa2eeaafe281fc5f58527256fbb53a58ea9020d65f684e674';

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
        url: 'https://bsc-dataseed1.binance.org/',
        chainId: 56,
        // blockNumber: 42163104, // block number of the latest block at the moment (without this, it gets error)
      },
    },
    bscTestnet: {
      url: 'https://data-seed-prebsc-1-s1.binance.org:8545/',
      chainId: 97,
      accounts: [
        PRIVATE_KEY_TEST,
      ],
      dex: {
        pancake: {
          factory: '0xB7926C0430Afb07AA7DEfDE6DA862aE0Bde767bc',
          router: '0x9Ac64Cc6e4415144C455BD8E4837Fea55603e5c3',
        },
        uniswapV2: {
          factory: '0x0000000000000000000000000000000000000000',
          router: '0x0000000000000000000000000000000000000000',
        },
      }
    },
    sepolia: {
      url: 'https://sepolia.infura.io/v3/dd7bf6951267477f8f20d852e0135f84',
      chainId: 11155111,
      accounts: [
        PRIVATE_KEY_TEST,
      ],
      dex: {
        // Only PancakeSwapV3 is supported on Sepolia
        pancake: {
          factory: '0x0000000000000000000000000000000000000000',
          router: '0x0000000000000000000000000000000000000000',
        },
        uniswapV2: {
          factory: '0x7E0987E5b3a30e3f2828572Bb659A548460a3003',
          router: '0xC532a74256D3Db42D0Bf7a0400fEFDbad7694008',
        },
      }
    },
    bscMainnet: {
      url: 'https://bsc-dataseed1.binance.org/',
      chainId: 56,
      accounts: [],
    },
  },
};
