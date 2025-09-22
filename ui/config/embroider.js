'use strict';

module.exports = function embroiderConfig() {
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

    // Performance?
    skipBabel: [
      {
        package: 'qunit',
      },
    ],
  };
};
