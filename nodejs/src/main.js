// https://ethereum.stackexchange.com/questions/144557/call-smart-contract-method-with-ethers-js-version-6

const ethers = require('ethers');
const fs = require('fs');
const QuoterABI = require('@uniswap/v3-periphery/artifacts/contracts/lens/Quoter.sol/Quoter.json').abi;
require('dotenv').config()

const provider = new ethers.JsonRpcProvider(`https://mainnet.infura.io/v3/${process.env.INFURA_API_KEY}`);

// get Uniswap V3 quote
const quoterAddress = '0xb27308f9F90D607463bb33eA1BeBb41C27CE5AB6';
const quoterContract = new ethers.Contract(quoterAddress, QuoterABI, provider);

// READ FILE /////////////////////////////////////////////////
function getFile(filePath) {
    try {
        return fs.readFileSync(filePath, 'utf8');
    } catch (e) {
        console.error('File read error:', e);
        return [];
    }
}

function calculateArb(amountIn, outputOut, surfaceResult) {
    const profitLoss = outputOut - amountIn;
    const profitLossPerc = (profitLoss / amountIn) * 100;

    return {
        profitLoss,
        profitLossPerc,
    };
}

// GET PRICE /////////////////////////////////////////////////
async function getPrice(address , amountIn, tradeDirection, verbose = true) {
    const ABI = [
        'function token0() external view returns (address)',
        'function token1() external view returns (address)',
        'function fee() external view returns (uint24)',
    ];
    const poolContract = new ethers.Contract(address, ABI, provider);
    const token0Address = await poolContract.token0();
    const token1Address = await poolContract.token1();
    const fee = await poolContract.fee();

    // Get individual token information (symbol, name, decimals)
    const addressArray = [token0Address, token1Address];
    const tokenInfo = [];

    for (let i = 0; i < addressArray.length; i++) {
        const tokenAddress = addressArray[i];
        const tokenABI = [
            'function symbol() external view returns (string)',
            'function name() external view returns (string)',
            'function decimals() external view returns (uint)',
        ];
        const tokenContract = new ethers.Contract(tokenAddress, tokenABI, provider);
        const tokenSymbol = await tokenContract.symbol();
        const tokenName = await tokenContract.name();
        const tokenDecimals = await tokenContract.decimals();
        tokenInfo.push({
            id: `token${i}`,
            tokenAddress,
            tokenSymbol,
            tokenName,
            tokenDecimals
        });
    }

    let inputTokenA;
    let inputDecimalA;
    let inputTokenB;
    let inputDecimalB;

    if (tradeDirection === 'baseToQuote') {
        inputTokenA = token0Address;
        inputDecimalA = tokenInfo[0].tokenDecimals;
        inputTokenB = token1Address;
        inputDecimalB = tokenInfo[1].tokenDecimals;
    } else if (tradeDirection === 'quoteToBase') {
        inputTokenA = token1Address;
        inputDecimalA = tokenInfo[1].tokenDecimals;
        inputTokenB = token0Address;
        inputDecimalB = tokenInfo[0].tokenDecimals;
    }

    // reformat amountIn
    let amountInParsed = ethers.parseUnits(amountIn.toString(), inputDecimalA);
    let quotedAmountOut = 0;

    try {
        quotedAmountOut = await quoterContract.quoteExactInputSingle.staticCall(
            inputTokenA,
            inputTokenB,
            fee,
            amountInParsed,
            0
        );
        quotedAmountOut = ethers.formatUnits(quotedAmountOut, inputDecimalB);

        return quotedAmountOut
    } catch (err) {
        if (verbose) console.error('Quoter error:', err);
        return 0
    }
}

// GET DEPTH /////////////////////////////////////////////////
async function getDepthForEachSurfaceOpportunity(surfaceResult, amountIn) {
    const pair1Contract = surfaceResult.contract1;
    const pair2Contract = surfaceResult.contract2;
    const pair3Contract = surfaceResult.contract3;
    const pair1ContractAddress = surfaceResult.contract1Address;
    const pair2ContractAddress = surfaceResult.contract2Address;
    const pair3ContractAddress = surfaceResult.contract3Address;
    const directionTrade1 = surfaceResult.directionTrade1;
    const directionTrade2 = surfaceResult.directionTrade2;
    const directionTrade3 = surfaceResult.directionTrade3;

    console.log('Checking trade 1 acquired coin...');
    const acquiredCoinT1 = await getPrice(pair1ContractAddress, 0.1, directionTrade1, verbose=false);

    // console.log('Checking trade 2 acquired coin...');
    if (acquiredCoinT1 === 0) return;
    const acquiredCoinT2 = await getPrice(pair2ContractAddress, acquiredCoinT1, directionTrade2, verbose=false);

    // console.log('Checking trade 3 acquired coin...');
    if (acquiredCoinT2 === 0) return;
    const acquiredCoinT3 = await getPrice(pair3ContractAddress, acquiredCoinT2, directionTrade3, verbose=false);

    // Calculate and show result
    const depthResult = calculateArb(amountIn, acquiredCoinT3, surfaceResult);

    if (depthResult.profitLossPerc >= 0) {
        console.log('Arbitrage opportunity:', depthResult, surfaceResult);
        console.log('Amount out:', acquiredCoinT3, pair1Contract, pair2Contract, pair3Contract);
    } else {
        console.log('No arbitrage opportunity for', pair1Contract, pair2Contract, pair3Contract);
    }
}

async function getDepth(amountIn, limit) {
    console.log('Reading surface rate information...');
    let fileInfo = getFile('../../golang/src/triangular_arb_surface_results.json');
    const fileJsonArray = JSON.parse(fileInfo);
    const promises = [];

    // fileJsonArray.forEach(surfaceResult => {
    //     promises.push(getDepthForEachSurfaceOpportunity(surfaceResult, amountIn));
    // })
    // await Promise.all(promises);
    for (let i = 0; i < fileJsonArray.length; i++) {
        await (getDepthForEachSurfaceOpportunity(fileJsonArray[i], amountIn));
    }
}

getDepth(amountIn=10).then(() => {
    console.log('Done');
});
