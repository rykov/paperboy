import { inject as service } from '@ember/service';
import { tracked } from '@glimmer/tracking';
import Controller from '@ember/controller';
import { isPresent } from '@ember/utils';
import { action } from '@ember/object';

export default class PreviewIndexController extends Controller {
  @service router;

  // Form fields
  @tracked listID = '';
  @tracked campaignID = '';

  // Process manual preview form
  @action startPreview() {
    const c = this.campaignID,
      l = this.listID;
    if (isPresent(c) && isPresent(l)) {
      this.router.transitionTo('preview.render', c, l);
    }
  }
}
