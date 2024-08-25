import express from 'express';

import UniswapService from './service.js';
import { batchDepthCalculationRules, depthCalculationRules, loadTokenRules } from './validation.js';
import { validateInputs } from '../common/validation.js';

const router = express.Router();
const service = new UniswapService();

router.post(
    '/arbitrage/depth',
    validateInputs(depthCalculationRules),
    async (req, res) => {
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
    }
);

router.post(
    '/arbitrage/batch-depth',
    validateInputs(batchDepthCalculationRules),
    async (req, res) => {
        const results = await service.getBatchDepthOpportunity(req.body.surfaceResults);
        console.log(JSON.stringify(req.body), results);

        return res.status(200).json(results);
    }
);

router.post(
    '/tokens/load',
    validateInputs(loadTokenRules),
    async (req, res) => {
        try {
            const result = await service.loadTokens(req.body.pairAddresses);

            return res.status(200).json(result);
        } catch(e) {
            return res.status(500).json({ error: e.message });
        }
    }
);

export default router;
