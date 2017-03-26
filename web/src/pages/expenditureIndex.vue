<template>
    <div>
        <h2 class="page-title">Uitgaves</h2>
        <h3 class="sub-title">{{title}}</h3>
        <div class="pure-g">
            <div class="pure-u-1 pure-u-md-3-4">
                <expenditures-table v-bind:expenditures="expenditures"
                                    v-on:expenditure-deleted="expenditureDeleted" />
            </div>
            <div class="pure-u-1 pure-u-md-1-4">
                <stats-chart v-bind:stats="stats"></stats-chart>
            </div>
        </div>
    </div>
</template>

<script lang="ts">
import Vue from "vue";
import moment from 'moment';

import StatsChart from '@/components/statsChart.vue';
import ExpendituresTable from '@/components/expendituresTable.vue';


import api from '@/api';

export default {
  name: 'expenditure-index',
  components: {StatsChart, ExpendituresTable},
  data () {
    return {
        type: '',
        title: '',
        stats: [],
        expenditures: [],
        start: null,
        end: null,
    };
  },
  mounted () {
      this.load();
  },

  methods: {
      loadCategoryStats() {
        let self = this;

        api.getCategoryStats({start: this.start, end: this.end})
        .then((stats) => {
            self.stats = stats;
        }).catch((reason) => {
            (<any>$).notify('Kon statistieken niet laden: ' + reason.message, 'warn');
        });
      },
      loadExpenditures() {
        let self = this;
        api.getExpenditures({start: this.start, end: this.end, sort: 'date', order: 'desc'})
        .then((data) => {
            self.expenditures = data.data;
        }).catch((reason) => {
            (<any>$).notify('Kon uitgaves niet laden: ' + reason.message, 'warn');
        });
      },
      load() {
        let self = this;

        let title = '';
        let start: moment.Moment;
        let end: moment.Moment;
        switch(this.$route.params.type) {
            case 'day':
            start = moment(this.$route.query.start).startOf('day');
            end = moment(start).add(1, 'days').startOf('day');
            title = start.format('LL');
            break;

            case 'range':
            start = moment(this.$route.query.start);
            end = moment(this.$route.query.end);

            title = start.format('LL') + ' - ' + end.format('LL');
            break;

            default:
                throw Error('Invalid type in category stats.');
        }

        this.title = title;
        this.start = start;
        this.end = end;

        this.loadCategoryStats();
        this.loadExpenditures();
      },

      expenditureDeleted(e) {
          this.loadCategoryStats();
      },
  },

  watch: {
      '$route': function() {
          this.load();
      },
  },
};
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style lang="scss" scoped>

</style>
