# OriginLang Docs

The OriginLang documentation site, built with [Astro Starlight](https://starlight.astro.build).

## Content structure

- English (default): `src/content/docs/`
- Chinese: `src/content/docs/zh/`

## Commands

```sh
npm install        # install dependencies
npm run dev        # start the dev server at localhost:4321
npm run build      # build the production site to ./dist/
npm run preview    # preview the build locally
```

Running the dev server in background mode:

```sh
npx astro dev --background
npx astro dev stop
npx astro dev status
```

## Writing docs

Add Markdown/MDX pages under `src/content/docs/` (English) or
`src/content/docs/zh/` (Chinese), then register them in the sidebar in
`astro.config.mjs`.

Useful references:

- [Starlight docs](https://starlight.astro.build/)
- [Astro content collections](https://docs.astro.build/en/guides/content-collections/)
- [Starlight internationalization](https://starlight.astro.build/guides/i18n/)