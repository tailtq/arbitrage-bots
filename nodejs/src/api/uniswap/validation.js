import { body } from 'express-validator';

export const loadTokenRules = [
    body('pairAddresses').isArray(),
];

export const depthCalculationRules = [
    body('surfaceResult.swap1').isString(),
    body('surfaceResult.swap2').isString(),
    body('surfaceResult.swap3').isString(),
    body('surfaceResult.contract1').isString(),
    body('surfaceResult.contract2').isString(),
    body('surfaceResult.contract3').isString(),
    body('surfaceResult.contract1Address').isString(),
    body('surfaceResult.contract2Address').isString(),
    body('surfaceResult.contract3Address').isString(),
    body('surfaceResult.directionTrade1').isString(),
    body('surfaceResult.directionTrade2').isString(),
    body('surfaceResult.directionTrade3').isString(),
    body('surfaceResult.startingAmount').isFloat(),
    body('surfaceResult.acquiredCoinT1').isFloat(),
    body('surfaceResult.acquiredCoinT2').isFloat(),
    body('surfaceResult.acquiredCoinT3').isFloat(),
    body('surfaceResult.swap1Rate').isFloat(),
    body('surfaceResult.swap2Rate').isFloat(),
    body('surfaceResult.swap3Rate').isFloat(),
    body('surfaceResult.profitLoss').isFloat(),
    body('surfaceResult.profitLossPerc').isFloat(),
];

export const batchDepthCalculationRules = [
    body('surfaceResults.*.swap1').isString(),
    body('surfaceResults.*.swap2').isString(),
    body('surfaceResults.*.swap3').isString(),
    body('surfaceResults.*.contract1').isString(),
    body('surfaceResults.*.contract2').isString(),
    body('surfaceResults.*.contract3').isString(),
    body('surfaceResults.*.contract1Address').isString(),
    body('surfaceResults.*.contract2Address').isString(),
    body('surfaceResults.*.contract3Address').isString(),
    body('surfaceResults.*.directionTrade1').isString(),
    body('surfaceResults.*.directionTrade2').isString(),
    body('surfaceResults.*.directionTrade3').isString(),
    body('surfaceResults.*.startingAmount').isFloat(),
    body('surfaceResults.*.acquiredCoinT1').isFloat(),
    body('surfaceResults.*.acquiredCoinT2').isFloat(),
    body('surfaceResults.*.acquiredCoinT3').isFloat(),
    body('surfaceResults.*.swap1Rate').isFloat(),
    body('surfaceResults.*.swap2Rate').isFloat(),
    body('surfaceResults.*.swap3Rate').isFloat(),
    body('surfaceResults.*.profitLoss').isFloat(),
    body('surfaceResults.*.profitLossPerc').isFloat(),
];