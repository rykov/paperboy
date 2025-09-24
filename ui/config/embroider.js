'use strict';

const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const EmberApp = require('ember-cli/lib/broccoli/ember-app');
const PostCSSConfig = require('./postcss.js');

module.exports = function embroiderConfig() {
  const isProd = EmberApp.env() === 'production';
  const sourceMapCommon = { sourceMap: !isProd };

  return {
    staticAddonTestSupportTrees: true,
    staticEmberSource: true,
    staticAddonTrees: true,
    staticModifiers: true,

    // TODO: Fix & enable
    // staticComponents: true,
    // staticHelpers: true,

    // Load uncommon routes separately
    // splitAtRoutes: ['orgs', 'my', 'user'],

    // Configure Webpack
    packagerOptions: {
      webpackConfig: {
        plugins: [
          new MiniCssExtractPlugin({
            ...(isProd && { filename: '[name].[contenthash].css' }),
          }),
        ],
        module: {
          rules: [
            {
              // CSS stylesheets
              test: /\.css$/i,
              use: [
                {
                  loader: 'postcss-loader',
                  options: {
                    postcssOptions: {
                      ...PostCSSConfig.embroiderCSS,
                      ...sourceMapCommon,
                    },
                  },
                },
              ],
            },
            {
              // SCSS stylesheets
              test: /\.scss$/i,
              use: [
                MiniCssExtractPlugin.loader,
                {
                  loader: 'css-loader',
                  options: {
                    modules: { auto: /\.m\.\w+$/i },
                    ...sourceMapCommon,
                  },
                },
                {
                  loader: 'postcss-loader',
                  options: {
                    sourceMap: !isProd,
                    postcssOptions: {
                      ...PostCSSConfig.embroiderSCSS,
                      ...sourceMapCommon,
                    },
                  },
                },
              ],
            },
          ],
        },
      },
    },

    // Performance?
    skipBabel: [
      {
        package: 'qunit',
      },
    ],
  };
};
