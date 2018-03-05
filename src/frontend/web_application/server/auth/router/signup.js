const proxy = require('express-http-proxy');
const { query } = require('../../api/lib/api');
const { getApiHost } = require('../../config');
const { authenticate } = require('../lib/cookie');
const { DEFAULT_REDIRECT } = require('../lib/redirect');

const CONTEXT_SAFE = 'safe';

const authenticateAfterSignup = async (req, res, next) => {
  const { username, password, device } = req.body;

  query({
    path: '/api/v1/authentications',
    method: 'post',
  }, {
    body: { username, password, device, context: CONTEXT_SAFE },
    success: async (user) => {
      await authenticate(res, { user });
      const redirect = req.query.redirect || DEFAULT_REDIRECT;
      res.redirect(redirect);
    },
    error: (err) => {
      const error = new Error(`Unable to automatically authenticate after signup: ${err.message}`);
      error.status = 502;
      next(error);
    },
  });
};

const createSignupRouting = (router) => {
  const target = getApiHost();

  router.post('/signup', proxy(target, {
    proxyReqPathResolver: () => '/api/v1/users',
    skipToNextHandlerFilter: proxyRes => proxyRes.statusCode === 200,
  }));
  router.post('/signup', authenticateAfterSignup);
};

module.exports = createSignupRouting;
