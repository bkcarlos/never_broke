# Deployment configuration

Use `compose.yml` for the application, PostgreSQL and the private MCP connection.
Build instructions and required backend variables are documented in
`../docs/FEATURE_DOCUMENTATION.md`.

## Keep personal configuration outside Git

- Copy `public-ingress.conf.example` to `public-ingress.conf` and replace the
  example hostname and interface address for your server. The real file is ignored.
- Keep actual hostnames, SSH targets, release notes and backup locations in
  `README.local.md` (ignored), or in private server documentation.
- Keep backend credentials in `deploy/.env` and source sync settings in
  `deploy/sync.env`. Both are ignored. Never put credentials in example files.
- Database backups, deployed frontend assets, certificates and private keys
  must remain outside the tracked source tree.

## Public filing footer

Set `VITE_ICP_FILING` in `frontend-vue/.env.production.local`, then build:

```sh
cd frontend-vue
pnpm install --frozen-lockfile
pnpm build
```

The actual value is local configuration ignored by Git. `.env.example` documents
its name without a personal value. With a value configured, the login and
registration views display a link to `https://beian.miit.gov.cn/`; an empty value
omits the link. The authenticated workspace does not render this footer.

`VITE_*` variables are compiled into public browser assets. Use them only for
public display values, never passwords, tokens or private keys. Build artifacts
are ignored; distribute them directly to the deployment server.

## Public gateway and certificates

The example gateway redirects HTTP to HTTPS and proxies requests to the local
web service at `127.0.0.1:80`. MCP remains private. Keep the app's ports bound to
loopback and preserve existing database volumes when changing domain routing.

Before first certificate issuance, load only the HTTP server block. Use Certbot
HTTP-01 webroot validation with the shared gateway's challenge directory, then
install the complete HTTPS configuration. Configure a periodic renewal service
that renews eligible certificates, validates Nginx configuration and reloads it.
Verify that service with a dry run and check its timer is enabled.

Back up the current gateway configuration and frontend before publishing. Check
HTTPS, HTTP redirects, static assets and unauthenticated API rejection afterward.
Never use `docker compose down -v` to change domain routing.
