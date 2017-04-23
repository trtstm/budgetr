<template>
      <table class="pure-table"
              style="width: 100%">
          <thead>
              <tr>
                  <th v-on:click="sortOrder = !sortOrder; sorter = sortName(sortOrder)">Categorie <i class="fa" v-bind:class="{'fa-sort-desc': sortOrder, 'fa-sort-asc': !sortOrder}" aria-hidden="true"></i></th>
                  <th v-on:click="sortOrder = !sortOrder; sorter = sortName(sortOrder)">Totaal <i class="fa" v-bind:class="{'fa-sort-desc': sortOrder, 'fa-sort-asc': !sortOrder}" aria-hidden="true"></i></th>
              </tr>
          </thead>
  
          <tbody>
              <tr v-for="stat in sortedStats">
                <td>{{stat.name}}</td>
                <td>{{stat.total}}</td>
              </tr>
          </tbody>
      </table>
</template>

<script lang="ts">
import Vue from "vue";
export default {
  name: "category-stats-table",
  props: ['stats'],
  data () {
    return {
      chart: null,
      sortOrder: false,
      sorter: null,
    };
  },
  computed: {
    sortedStats () {
      return this.stats.slice().sort(this.sorter);
    },
  },
  mounted () {
    let self = this;
    self.sorter = self.sortName(false);
    this.$watch('stats', () => {
    });


  },

  methods: {
    sortName: function(reverse: boolean): any {
      return function(a, b) {
        let m = 1;
        if(reverse) {
          m = -1;
        }

        return m * a.name.localeCompare(b.name);
      };
    },

    sortTotal: function(reverse: boolean): any {
      return function(a, b) {
        let m = 1;
        if(reverse) {
          m = -1;
        }

        return m * (a.total-b.total);
      };
    },
  },
};
</script>

<style lang="scss" scoped>
th {
  user-select: none;
  cursor: pointer;
}
</style>
