This directory is a copy of the `../modules` directory with the following changes:

The following commands were executed:

```bash
npm i --save-global esbuild
npx esbuild **/*.js --allow-overwrite --outdir=. --define:process.env.NODE_ENV=\"production\"
```
