import EmberRouter from '@ember/routing/router';
import config from 'preview/config/environment';

export default class Router extends EmberRouter {
  location = config.locationType;
  rootURL = config.rootURL;
}

Router.map(function () {
  this.route('preview', function () {
    this.route('render', { path: '/:content_id/:list_id' });
  });
});
