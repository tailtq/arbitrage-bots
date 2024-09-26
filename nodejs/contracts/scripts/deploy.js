const { ethers, network } = require('hardhat');

async function main() {
  const [deployer] = await ethers.getSigners();
  console.log('Deploying contracts with the account:', deployer.address);
  console.log('Account balance:', (await deployer.getBalance()).toString());
  const { pancake, uniswapV2 } = network.config.dex;

  const Contract = await ethers.getContractFactory('ArbitrageExecutor');
  const contract = await Contract.deploy(
    pancake.factory, pancake.router, uniswapV2.factory, uniswapV2.router,
  );
  await contract.deployed();
  console.log('Contract address:', contract.address);
}

main()
  .then(() => process.exit(0))
  .catch(error => {
    console.error(error);
    process.exit(1);
  });