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
        const depthResult = {
            forward: resultForward,
            backward: resultBackward,
        };
        service.logArbOpportunity({ surfaceResult, depthResult });

        return res.status(200).json(depthResult);
    }
);

router.post(
    '/arbitrage/batch-depth',
    validateInputs(batchDepthCalculationRules),
    async (req, res) => {
        const results = await service.getBatchDepthOpportunity(req.body.surfaceResults);

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
