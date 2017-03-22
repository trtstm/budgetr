import Vue from "vue";
import Router from "vue-router";
Vue.use(Router);
import Hello from "@/components/hello";
export default new Router({
    routes: [
        {
            path: '/',
            name: 'Hello',
            component: Hello,
        },
    ],
});
//# sourceMappingURL=index.js.map