<template>
    <div>
        <h2 class="page-title">Categorieën</h2>
    
        <div style="position: relative;">
            <div v-if="loading"
                 class="loader">
                <i class="fa fa-spinner fa-spin spinner"></i>
            </div>
            <div class="pure-g">
                <div class="pure-u-1">
                    <table class="pure-table"
                           style="width: 100%">
                        <thead>
                            <tr>
                                <th>Categorie</th>
                                <th> </th>
                            </tr>
                        </thead>
    
                        <tbody>
                            <tr v-for="c in categories">
                                <td>
                                    <input class="full-width"
                                           v-bind:data-category="c.getId()"
                                           v-on:keyup.enter="save(c)"
                                           type="text"
                                           v-bind:value="c.getName()">
                                </td>
                                <td><a class="save-btn"
                                       v-on:click="save(c)"><i class="fa fa-fw fa-floppy-o"></i></a></td>
                            </tr>
                        </tbody>
                    </table>
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import Vue from "vue";

import api from '@/api';

export default {
  name: 'category-index',
  components: {},
  data () {
    return {
        categories: [],
        loading: false,
    };
  },
  mounted () {
      this.load();
  },

  methods: {
      load() {
        let self = this;
        api.getCategories()
        .then((categories) => {
            self.categories = categories.data;
        });
      },

      save(c) {
          let self = this;
          let name = (<any>document.querySelectorAll("[data-category='" + c.getId() + "']")[0]).value;
          if(name.length === 0) {
              (<any>$).notify('Categorie kan niet leeg zijn.', 'warn');
              self.$forceUpdate();
              return;
          }
          let oldName = c.getName();
          c.setName(name);

          self.loading = true;
          api.updateCategory(c)
          .then((c2) => {
              c.setName(c2.getName());
              self.$forceUpdate();
              (<any>$).notify('Categorie gewijzigd.', 'success');
              self.loading = false;
          })
          .catch((reason) => {
                (<any>$).notify('Kon categorie niet wijzigen: ' + reason.message, 'warn');
                c.setName(oldName);
                self.$forceUpdate();
                self.loading = false;
          });
      },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style lang="scss" scoped>
.save-btn {
    cursor: pointer;
}

.full-width {
    width: 100%;
}
</style>
