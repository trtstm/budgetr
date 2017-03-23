import Vue from "vue";
import Router from "vue-router";

Vue.use(Router);

import CategoryStats from "@/components/categoryStats";

export default new Router({
  routes: [
    {
      path: '/',
      name: 'CategoryStats',
      component: CategoryStats,
    },
  ],
});

