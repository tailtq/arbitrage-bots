// cannot use import dotenv from 'dotenv'; dotenv.config(); because imports are hoisted
// https://github.com/motdotla/dotenv/issues/206
import 'dotenv/config';
import express from 'express';
import morgan from 'morgan';
import uniswapRouter from './api/uniswap/views.js';

BigInt.prototype.toJSON = function () {
    return this.toString();
};

const app = express();
app.use(express.json());
app.use(morgan('combined'))
app.use('/uniswap', uniswapRouter);

app.get('/', function (req, res) {
    res.status(200).send('Hello World');
});

app.listen(3000, function () {
    console.log('Server is running on port 3000');
});
