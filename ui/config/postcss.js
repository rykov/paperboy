// Configuration guide based on the following:
// https://github.com/jeffjewiss/ember-cli-postcss

// Packages installed for PostCSS
// - @csstools/postcss-sass
// - ember-cli-postcss
// - postcss-scss

const Browsers = require('./targets.js').browsers;
const tailwindConfig = './config/tailwind.js';

module.exports = function (/* isProduction */) {
  return {
    compile: {
      parser: require('postcss-scss'),
      extension: 'scss',
      plugins: [
        {
          module: require('@csstools/postcss-sass'),
          options: {
            includePaths: ['node_modules'],
          },
        },
        require('tailwindcss')(tailwindConfig),
        require('autoprefixer')({
          overrideBrowserslist: Browsers,
        }),
      ],
    },
  };
};
