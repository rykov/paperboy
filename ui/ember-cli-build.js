'use strict';

const EmberApp = require('ember-cli/lib/broccoli/ember-app');
const ConfigurePostCSS = require('./config/postcss.js');

module.exports = function (defaults) {
  const isProduction = EmberApp.env() === 'production';
  const app = new EmberApp(defaults, {
    // ember-cli-postcss
    postcssOptions: ConfigurePostCSS(),
  });

  // HACK: Ember 6.1+ removed "ember" import, so we do it here
  const emberJs = `ember.${isProduction ? 'prod' : 'debug'}.js`;
  app.import(`node_modules/ember-source/dist/${emberJs}`);

  // Use `app.import` to add additional libraries to the generated
  // output files.
  //
  // If you need to use different assets in different
  // environments, specify an object as the first parameter. That
  // object's keys should be the environment name and the values
  // should be the asset to use in that environment.
  //
  // If the library that you are including contains AMD or ES6
  // modules that you would like to import into your application
  // please specify an object with the list of modules as keys
  // along with the exports of each module as its value.

  const { Webpack } = require('@embroider/webpack');
  const emOpts = require('./config/embroider.js')();
  return require('@embroider/compat').compatBuild(app, Webpack, emOpts);
};
