import express from 'express';
import { body, validationResult } from 'express-validator';

import UniswapService from './service.js';

const router = express.Router();
const service = new UniswapService();

router.post('/arbitrage/depth', [
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
    body('surfaceResult.direction').isString(),
    body('surfaceResult.tradeDescription1').isString(),
    body('surfaceResult.tradeDescription2').isString(),
    body('surfaceResult.tradeDescription3').isString()
], async (req, res) => {
    const errors = validationResult(req);

    if (!errors.isEmpty()) {
        return res.status(400).json({ errors: errors.array() });
    }

    const { surfaceResult } = req.body;
    const [resultForward, resultBackward] = await Promise.all([
        service.getDepthOpportunityForward(surfaceResult),
        service.getDepthOpportunityBackward(surfaceResult),
    ]);
    console.log(JSON.stringify(req.body), resultForward, resultBackward);

    return res.status(200).json({
        forward: resultForward,
        backward: resultBackward
    });
});

router.post('/tokens/load', [
    body('pairAddresses').isArray()
], async (req, res) => {
    try {
        const errors = validationResult(req);

        if (!errors.isEmpty()) {
            return res.status(400).json({ errors: errors.array() });
        }

        const result = await service.loadTokens(req.body.pairAddresses);

        return res.status(200).json(result);
    } catch(e) {
        return res.status(500).json({ error: e.message });
    }
});

export default router;
