const express = require('express');
require('dotenv').config();

const uniswapRouter = require('./api/uniswap/views');

BigInt.prototype.toJSON = function () {
    return this.toString();
};

const app = express();
app.use(express.json());
app.use('/uniswap', uniswapRouter);

app.get('/', function (req, res) {
    res.status(200).send('Hello World');
});

app.listen(3000, function () {
    console.log('Server is running on port 3000');
});
