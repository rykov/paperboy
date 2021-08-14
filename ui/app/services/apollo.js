import ApolloService from 'ember-apollo-client/services/apollo';

const defaultOptions = {
  watchQuery: {
    fetchPolicy: 'network-only',
    errorPolicy: 'none',
  },
  query: {
    fetchPolicy: 'network-only',
    errorPolicy: 'none',
  },
  mutate: {
    errorPolicy: 'none',
  },
};

export default class PreviewApolloService extends ApolloService {
  clientOptions() {
    const opts = super.clientOptions();
    return Object.assign({}, opts, { defaultOptions });
  }
}
