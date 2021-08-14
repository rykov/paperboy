import query from 'preview/gql/queries/preview.graphql';
import { queryManager } from 'ember-apollo-client';
import Route from '@ember/routing/route';

export default class PreviewRenderRoute extends Route {
  @queryManager apollo;

  model(params) {
    let c = params.content_id;
    let r = params.list_id + '#0';
    return this.apollo.query(
      {
        fetchPolicy: 'network-only', // no cache
        variables: { recipient: r, content: c },
        query,
      },
      'renderOne'
    );
  }
}
