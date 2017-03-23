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

  </div>
</template>

<script lang="ts">
import Vue from "vue";
import moment from 'moment';
import * as $ from 'jquery';
import '@/static/vendor/bootstrap-daterangepicker/daterangepicker.js';
import '@/static/vendor/bootstrap-daterangepicker/daterangepicker.css';

import api from '@/api';

interface CategoryStats extends Vue {
  toDayView(day:moment.Moment) : void;
  toWeekView(week:moment.Moment) : void;
  toMonthView(month:moment.Moment) : void;
  start: moment.Moment;
  end: moment.Moment;
  stats: Array<any>,
  title: string,
}

export default {
  name: "category-stats",
  data () {
    return {
      start: null,
      end: null,
      stats: [],
      title: '',
    };
  },

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
    toDayView (day: moment.Moment) {
      let self = this;
      self.title = day.format('LL');
      api.getCategoryStats({start: day, end: moment(day).add(1, 'day').startOf('day')})
      .then((stats) => {
        self.stats = stats;
      })
    },
    toWeekView (week: moment.Moment) {
      let self = this;
      let end = moment(week).add(1, 'week').startOf('week');

      self.title = 'Van ' + week.format('LL') + ' tot ' + end.format('LL');
      api.getCategoryStats({start: week, end: end})
      .then((stats) => {
        self.stats = stats;
      })
    },
    toMonthView (month: moment.Moment) {
      let self = this;
      self.title = month.format('MMMM, YYYY');
      api.getCategoryStats({start: month, end: moment(month).add(1, 'month').startOf('month')})
      .then((stats) => {
        self.stats = stats;
      })
    },
  },
} as Vue.ComponentOptions<CategoryStats>;
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
