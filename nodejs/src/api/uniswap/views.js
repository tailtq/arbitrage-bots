import express from 'express';

import UniswapService from './services/uniswap.js';
import {
    batchDepthCalculationRules, depthCalculationRules, loadTokenRules, tokenPairPriceRules,
} from './validation.js';
import { validateInputs } from '../common/validation.js';
import TokenPriceAggregationService from './services/tokenPriceAggregation.js';

const router = express.Router();
const uniswapService = new UniswapService();
const tokenPriceAggService = new TokenPriceAggregationService(uniswapService, 10000, 5);
// tokenPriceAggService.startPricePolling();

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
        try {
            const results = await uniswapService.getBatchDepthOpportunity(req.body.surfaceResults);
            return res.status(200).json(results);
        } catch (e) {
            return res.status(500);
        }
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

router.post('/price/depth', validateInputs(tokenPairPriceRules), (req, res) => {
    const { tokenPairs } = req.body;
    const result = tokenPriceAggService.getPairPrices(tokenPairs);
    return res.status(200).json(result);
});

export default router;
