<template>
  <div>
    <h2 class="page-title">Budgetr</h2>
  
    <form class="pure-form pure-form-stacked main-form"
          v-on:submit.prevent="submitExpenditure()">
      <div class="pure-g">
        <div class="pure-u-1-2 pure-u-lg-1-2">
          <label for="first-name">Uitgave</label>
          <input ref="amountInput"
                 class="pure-u-1"
                 type="text">
        </div>
  
        <div class="pure-u-1-2 pure-u-lg-1-2">
          <label for="last-name">Category</label>
          <select ref="categoryInput"
          placeholder="Kies een categorie"
                  class="pure-u-1">
          </select>
        </div>
      </div>
  
      <div class="pure-g">
        <div class="pure-u-1 pure-u-lg-1">
          <input type="submit"
                 class="pure-u-1 pure-button pure-button-primary"
                 value="Toevoegen"
                 :disabled="submitting">
        </div>
      </div>
  
    </form>
  
    <h3 class="sub-title">Recent toegevoegd</h3>
    <expenditures-table format="fromNow"
                        v-bind:expenditures="buffer"
                        v-bind:limit="historyLength"
                        v-on:expenditure-deleted="expenditureDeleted" />
  </div>
</template>

<script lang="ts">
import Vue from "vue";

import moment from 'moment';
import {Parser} from 'expr-eval';

import ExpendituresTable from '@/components/expendituresTable.vue';

import api from '@/api';
import Expenditure from '@/expenditure';
import Category from '@/category';


export default {
    name: "home",
    data: () => {
        return {
            submitting: false,
            buffer: [],
            historyLength: 5,
        };
    },

    components: {ExpendituresTable},

    mounted() {
      let self = this;

        api.getExpenditures({sort:'date', order:'desc', start: moment(), end: moment().add(-1, 'days')})
        .then((data) => {
          //this.loaded = true;
          //self.latestExpenditures = data.data;
          
        }).catch((reason) => {

          (<any>$).notify('Kon de laatste uitgaves niet laden: ' + reason.message, 'warn');
          //this.loaded = true;
        });

        api.getCategories()
        .then((data) => {
          self.$refs.categoryInput.selectize.clearOptions();
          self.$refs.categoryInput.selectize.addOption(data.data);
          self.$refs.categoryInput.selectize.refreshOptions(false);
        }).catch((reason) => {
          (<any>$).notify('Kon geen categorieën laden: ' + reason.message, 'warn');
        });

        this.loadExpenditures();

          $(self.$refs.categoryInput).selectize({
            create: true,
            labelField: 'name',
            valueField: 'name',
            sortField: {
              field: 'name',
              direction: 'asc',
            },
          });
    },
    

    methods: {
      loadExpenditures() {
        let self = this;
        api.getExpenditures({sort: 'date', order: 'desc', limit: self.historyLength*2})
        .then((data) => {
          self.buffer = data.data;
        }).catch((reason) => {
          (<any>$).notify('Kon geen categorieën laden: ' + reason.message, 'warn');
        });
      },
        submitExpenditure() {
            let self = this;
            this.submitting = true;

            let amount = 0;
            try {
              amount = Parser.evaluate(this.$refs.amountInput.value.replace(/,/g, '.'));
            } catch(err) {

            }
            let category = this.$refs.categoryInput.value;
            if(category.length === 0) {
              this.submitting = false;
              return;
            }

            let expenditure = new Expenditure({ amount: amount, date: moment() });
            expenditure.setCategory(new Category({ name: category }));

            api.createExpenditure(expenditure)
                .then((e) => {
                    self.submitting = false;

                    self.buffer.unshift(e);
                    if(expenditure.getCategory()) {
                      self.$refs.categoryInput.selectize.addOption(expenditure.getCategory());
                    }

                    self.$refs.amountInput.value = '';
                    self.$refs.amountInput.focus();

                    (<any>$).notify('Uitgave toegevoegd.', 'success');
                }).catch((reason) => {
                    self.submitting = false;

                    (<any>$).notify(reason.message, 'warn');
                });
        },
        expenditureDeleted(e) {
          // Check if our new length is below treshhold.
          if(this.buffer.length < 1.25 * this.historyLength) {
            this.loadExpenditures();
          }
        },
    },
    computed: {

    },
};
</script>

<style lang="scss">
.content {
  max-width: 768px;
  margin-left: auto;
  margin-right: auto;
}

.title {
  text-align: center;
}

.main-form {
  border-bottom: 1px solid #eee;
  padding-bottom: 1em;
  margin-bottom: 1em;
}

.latest-expenditures {
  margin-left: auto;
  margin-right: auto;
}
</style>

