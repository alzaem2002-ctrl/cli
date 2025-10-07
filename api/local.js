const express = require('express');
const app = express();

app.get('/', (req, res) => {
  res.status(200).json({ status: 'ok', service: 'cli-backend' });
});

app.get('/health', (req, res) => {
  res.status(200).json({ ok: true });
});

const port = process.env.PORT || 3000;
app.listen(port, () => {
  console.log(`Server listening on port ${port}`);
});
