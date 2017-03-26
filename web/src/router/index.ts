import Vue from "vue";
import Router from "vue-router";

Vue.use(Router);

import Home from "@/pages/home.vue";
import ExpenditureIndex from "@/pages/expenditureIndex.vue";


export default new Router({
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home,
    },
    {
      path: '/expenditure-index/:type',
      name: 'expenditure-index',
      component: ExpenditureIndex,
    },
  ],
});

