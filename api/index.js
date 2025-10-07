const express = require('express');
const app = express();

app.get('/', (req, res) => {
  res.status(200).send('✅ Backend is Live!');
});

app.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok' });
});

module.exports = (req, res) => app(req, res);
