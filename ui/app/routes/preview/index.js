import query from 'preview/gql/queries/index.graphql';
import { queryManager } from 'ember-apollo-client';
import Route from '@ember/routing/route';

export default class PreviewIndexRoute extends Route {
  @queryManager apollo;

  model() {
    return this.apollo.query({
      query,
    });
  }
}
