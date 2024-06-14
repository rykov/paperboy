import { inject as service } from '@ember/service';
import { tracked } from '@glimmer/tracking';
import Controller from '@ember/controller';
import { isPresent } from '@ember/utils';
import { action } from '@ember/object';

export default class PreviewIndexController extends Controller {
  @service router;

  // Form fields
  @tracked selectedList;
  @tracked selectedCampaign;

  // Process manual preview form
  @action startPreview() {
    const l = this.selectedList?.param;
    const c = this.selectedCampaign?.param;
    if (isPresent(c) && isPresent(l)) {
      this.router.transitionTo('preview.render', c, l);
    }
  }
}
