const express = require('express');
const cors = require('cors');

const app = express();
const PORT = process.env.PORT || 3000;

// Middleware
app.use(cors());
app.use(express.json());

// Track start time for uptime calculation
const startTime = Date.now();

// Routes
app.get('/', (req, res) => {
  res.send('Backend API is running successfully!');
});

app.get('/health', (req, res) => {
  res.json({ status: 'ok' });
});

app.get('/status', (req, res) => {
  const uptime = Math.floor((Date.now() - startTime) / 1000);
  res.json({
    status: 'running',
    uptime: uptime,
    timestamp: new Date().toISOString(),
    environment: process.env.NODE_ENV || 'production'
  });
});

// Start server
app.listen(PORT, () => {
  console.log(`Server running on port ${PORT}`);
});