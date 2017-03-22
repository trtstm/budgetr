import Vue from "vue";
import router from "./router";

import App from "./app";

import Api from './api';


new Vue({
  el: "#app",
  router,
  render: h => h(App),
});
