<template>
  <div id="app"
       class="">
    <div class="pure-menu pure-menu-horizontal nav-menu">
      <ul class="pure-menu-list">
        <li class="pure-menu-item pure-menu-selected">
          <router-link class="pure-menu-link"
                       :to="{name: 'home'}">Home</router-link>
        </li>
        <li class="pure-menu-item pure-menu-has-children pure-menu-allow-hover">
          <a class="pure-menu-link">Uitgaves</a>
          <ul class="pure-menu-children">
            <li class="pure-menu-item">
              <router-link class="pure-menu-link"
                           :to="{name: 'expenditure-index', params: {type: 'day'}, query: {start: now.format()}}">Vandaag</router-link>
            </li>
            <li class="pure-menu-item">
              <router-link class="pure-menu-link"
                           :to="{name: 'expenditure-index', params: {type: 'range'}, query: {start: moment(now).startOf('week').format(), end: moment(now).endOf('week').format()}}">Deze week</router-link>
            </li>
            <li class="pure-menu-item">
              <router-link class="pure-menu-link"
                           :to="{name: 'expenditure-index', params: {type: 'range'}, query: {start: moment(now).startOf('month').format(), end: moment(now).endOf('month').format()}}">Deze maand</router-link>
            </li>
  
            <li class="pure-menu-item">
              <a href="#"
                 v-on:click.prevent
                 ref="dayPicker"
                 class="pure-menu-link">Kies een dag</a>
            </li>
            <li class="pure-menu-item">
              <a href="#"
                 v-on:click.prevent
                 ref="rangePicker"
                 class="pure-menu-link">Kies een bereik</a>
            </li>
          </ul>
        </li>
  
        <li class="pure-menu-item pure-menu-has-children pure-menu-allow-hover">
          <a class="pure-menu-link">Exporteer</a>
          <ul class="pure-menu-children">
            <li class="pure-menu-item">
              Laatste
              <input ref="exportNumber"
                     type="number">
              <select ref="exportType">
                <option value="day">Dagen</option>
                <option value="week">Weken</option>
                <option value="month">Maanden</option>
              </select>
              <button v-on:click="createExport">Exporteer</button>
            </li>
          </ul>
        </li>
      </ul>
    </div>
  
    <transition name="fade"
                mode="out-in">
      <router-view class="content"></router-view>
    </transition>
  </div>
</template>

<script lang="ts">
import Vue from "vue";

import 'selectize';
import moment from 'moment';
import '../node_modules/selectize/dist/css/selectize.css';

import '../node_modules/font-awesome/css/font-awesome.css';

import Flatpickr from 'flatpickr'
import 'flatpickr/dist/themes/airbnb.css'

import api from '@/api';


export default {
  name: "app",
  data () {
    return {
      now: moment(),
      moment: moment,
    }
  },

  mounted() {
    let self = this;
    new Flatpickr(this.$refs.dayPicker, {
      disableMobile: true,
      onChange(dates) {
        self.$router.push({name: 'expenditure-index', params: {type: 'day'}, query: {start: moment(dates[0]).format()}});
      },
    });

    new Flatpickr(this.$refs.rangePicker, {
      disableMobile: true,
      mode: 'range',
      onChange(dates) {
        if(dates.length < 2) return;
  
        self.$router.push({name: 'expenditure-index', params: {type: 'range'}, query: {start: moment(dates[0]).format(), end: moment(dates[1]).format()}});
      },
    });
  },

  methods: {
    createExport() {
      let n = Number(this.$refs.exportNumber.value);
      let type = this.$refs.exportType.value;

      let ranges : any = [];
      ranges.push({
        start: moment().startOf(type),
        end: moment(),
        title: moment().startOf(type).format('DD/MM/YYYY'),
      });

      for(let i = 1; i < n+1; i++) {
        ranges.push({
          start: moment().add(-i, type).startOf(type),
          end: moment().add(-i, type).endOf(type),
          title: moment().add(-i, type).startOf(type).format('DD/MM/YYYY'),
        });
      }

      ranges.forEach((r) => {
        console.log('start: ' + r.start.format());
        console.log('end: ' + r.end.format());
      });

      api.generateExcel(ranges);

    },
  },
};
</script>

<style src="./assets/sass/style.scss" lang="scss"></style>

