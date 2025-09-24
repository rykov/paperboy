// Configuration guide based on the following:
// https://github.com/jeffjewiss/ember-cli-postcss

// Packages installed for PostCSS
// - @csstools/postcss-sass
// - ember-cli-postcss
// - postcss-scss

const Browsers = require('./targets.js').browsers;

// Common plugins for CSS & SCSS
const commonPlugins = [
  require('@tailwindcss/postcss')({}),
  require('autoprefixer')({
    overrideBrowserslist: Browsers,
  }),
];

module.exports = {
  // Configuration for CSS pipeline
  embroiderCSS: {
    plugins: commonPlugins,
  },

  // Configuration for SCSS pipeline
  embroiderSCSS: {
    parser: require('postcss-scss'),
    extension: 'scss',
    plugins: [
      require('@csstools/postcss-sass')({
        includePaths: ['node_modules'],
      }),
      ...commonPlugins,
    ],
  },
};
