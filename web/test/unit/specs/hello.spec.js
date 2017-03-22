import { expect } from "chai";
import Vue from "vue";
import Hello from "@/components/hello";
describe("Hello.vue", function () {
    it("should render correct contents", function () {
        var vm = new Vue({
            el: document.createElement("div"),
            render: function (h) { return h(Hello); },
        });
        var subtitle = vm.$el.querySelector("subtitle");
        if (subtitle !== null) {
            expect(subtitle.textContent).to.equal("Welcome to Your Vue.js App");
        }
    });
});
//# sourceMappingURL=hello.spec.js.map