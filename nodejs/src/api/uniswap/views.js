import express from 'express';

import UniswapService from './services/uniswap.js';
import { batchDepthCalculationRules, depthCalculationRules, loadTokenRules } from './validation.js';
import { validateInputs } from '../common/validation.js';
import TokenPriceAggregationService from './services/tokenPriceAggregation.js';

const router = express.Router();
const uniswapService = new UniswapService();
new TokenPriceAggregationService(uniswapService, 60000).startPricePolling();

router.post(
    '/arbitrage/depth',
    validateInputs(depthCalculationRules),
    async (req, res) => {
        const { surfaceResult } = req.body;
        const [resultForward, resultBackward] = await Promise.all([
            uniswapService.getDepthOpportunityForward(surfaceResult),
            uniswapService.getDepthOpportunityBackward(surfaceResult),
        ]);
        const depthResult = {
            forward: resultForward,
            backward: resultBackward,
        };
        uniswapService.logArbOpportunity({ surfaceResult, depthResult });

        return res.status(200).json(depthResult);
    }
);

router.post(
    '/arbitrage/batch-depth',
    validateInputs(batchDepthCalculationRules),
    async (req, res) => {
        console.log(req.body.surfaceResults);
        const results = await uniswapService.getBatchDepthOpportunity(req.body.surfaceResults);

        return res.status(200).json(results);
    }
);

router.post(
    '/tokens/load',
    validateInputs(loadTokenRules),
    async (req, res) => {
        try {
            const result = await uniswapService.loadTokens(req.body.pairAddresses);

            return res.status(200).json(result);
        } catch(e) {
            return res.status(500).json({ error: e.message });
        }
    }
);

export default router;
