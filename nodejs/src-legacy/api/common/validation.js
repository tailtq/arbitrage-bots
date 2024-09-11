// nodejs/src/api/uniswap/validation.js
import { validationResult } from 'express-validator';

export const validateInputs = (inputs) => {
    const validateFunc = (req, res, next) => {
        const errors = validationResult(req);
        if (!errors.isEmpty()) {
            return res.status(400).json({ errors: errors.array() });
        }
        next();
    };
    return [inputs, validateFunc];
};
