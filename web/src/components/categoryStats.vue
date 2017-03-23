<template>
  <div>
    <h3>{{title}}</h3>

    <a href="#" ref="dayPicker">Toon dag</a>
    <a href="#" ref="weekPicker">Toon week</a>
    <a href="#" ref="monthPicker">Toon maand</a>

    <ul>
      <li v-for="stat in stats">
        {{stat.name}}: {{stat.total}}
      </li>
    </ul>

    <div class="row">
      <div v-for="stat in stats" class="col-md-2">
        <div class="square" v-bind:style="{'background-color': stat.color}">
          <span class="stat-info">{{stat.name}}</span>
        </div>
      </div>
    </div>

    <expenditures-table v-bind:expenditures="expenditures"></expenditures-table>

  </div>
</template>

<script lang="ts">
import Vue from "vue";
import moment from 'moment';
import * as $ from 'jquery';
import * as _ from 'lodash';
import '@/static/vendor/bootstrap-daterangepicker/daterangepicker.js';
import '@/static/vendor/bootstrap-daterangepicker/daterangepicker.css';

import expendituresTable from './expendituresTable.vue';

import api from '@/api';

interface CategoryStats extends Vue {
  initStats(stats:any) : void;
  toDayView(day:moment.Moment) : void;
  toWeekView(week:moment.Moment) : void;
  toMonthView(month:moment.Moment) : void;
  start: moment.Moment;
  end: moment.Moment;
  stats: Array<any>;
  title: string;
  expenditures: Array<any>;
}

export default {
  name: "category-stats",
  data () {
    return {
      start: null,
      end: null,
      stats: [],
      title: '',
      expenditures: [],
    };
  },
  components: {expendituresTable},
  mounted () {
    var self = this;

    $(this.$refs.dayPicker).daterangepicker({
        singleDatePicker: true,
        showDropdowns: true,
    }, (start: any, end: any, label: any) => {
      this.toDayView(start);
    });

    $(this.$refs.weekPicker).daterangepicker({
        singleDatePicker: true,
        showDropdowns: true,
    }, (start: any, end: any, label: any) => {
      this.toWeekView(start.startOf('week'));
    });

    $(this.$refs.monthPicker).daterangepicker({
        singleDatePicker: true,
        showDropdowns: true,
    }, (start: any, end: any, label: any) => {
      this.toMonthView(start.startOf('month'));
    });

    this.toMonthView(moment().startOf('month'));
  },

  methods: {
    initStats (stats: any) {
        for(let i = 0; i < stats.length; i++) {
          let h = i * 360/stats.length;
          let s = 100;
          let l = 50;
          stats[i].color = 'hsl(' + h + ', ' + s + '%, ' + l + '%)';
        }

        this.stats = stats;
    },
    toDayView (day: moment.Moment) {
      let self = this;
      self.title = day.format('LL');
      let end = moment(day).add(1, 'days').startOf('day');
      api.getCategoryStats({start: day, end: end})
      .then((stats) => {
        self.initStats(stats);
        api.getExpenditures({start: day, end: end, sort: 'date', order:'desc'})
        .then((expenditures) => {
          self.expenditures = expenditures.data;
        });
      })
    },
    toWeekView (week: moment.Moment) {
      let self = this;
      let end = moment(week).add(1, 'week').startOf('week');

      self.title = 'Van ' + week.format('LL') + ' tot ' + end.format('LL');
      api.getCategoryStats({start: week, end: end})
      .then((stats) => {
        self.initStats(stats);
        api.getExpenditures({start: week, end: end, sort: 'date', order:'desc'})
        .then((expenditures) => {
          self.expenditures = expenditures.data;
        });
      })
    },
    toMonthView (month: moment.Moment) {
      let self = this;
      let end = moment(month).add(1, 'month').startOf('month');
      self.title = month.format('MMMM, YYYY');
      api.getCategoryStats({start: month, end: end})
      .then((stats) => {
        self.initStats(stats);
        api.getExpenditures({start: month, end: end, sort: 'date', order:'desc'})
        .then((expenditures) => {
          self.expenditures = expenditures.data;
        });
      })
    },
  },
} as Vue.ComponentOptions<CategoryStats>;
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style lang="scss" scoped>
 .square {
   width: 100%;
   padding-top: 100%;
   position: relative;

   text-align: center;

   .stat-info {
     position: absolute;
     top: 50%;
   }
 }
</style>
