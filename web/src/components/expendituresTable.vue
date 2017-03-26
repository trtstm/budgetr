<template>
    <div style="position: relative;">
        <div v-if="loading"
             class="loader">
            <i class="fa fa-spinner fa-spin spinner"></i>
        </div>
        <table class="pure-table"
               style="width: 100%">
            <thead>
                <tr>
                    <th>Datum</th>
                    <th>Uitgave</th>
                    <th>Categorie</th>
                    <th> </th>
                </tr>
            </thead>
    
            <tbody>
                <tr v-for="e in capped">
                    <td>{{formatDate(e.getDate())}}</td>
                    <td>{{e.getAmount()}}</td>
                    <td>{{categoryName(e.getCategory())}}</td>
                    <td><a v-on:click="deleteExpenditure(e)"><i class="fa fa-fw fa-trash-o"></i></a></td>
                </tr>
            </tbody>
        </table>
    </div>
</template>

<script lang="ts">
import Vue from "vue";
import moment from 'moment';

import api from '@/api';
import Category from '@/category';

export default {
    name: "expenditures-table",
    props: {expenditures: {default: []}, limit: {default: -1}, format: {default: 'LL'}},
    data () {
        return {
            loading: false,
        };
    },
    methods: {
        categoryName (category: Category | null) {
            if(category === null) {
                return '';
            }

            return category.getName();
        },

        deleteExpenditure(e) {
            let self = this;
            self.loading = true;
            api.deleteExpenditure(e)
                .then(() => {
                    (<any>$).notify('Uitgave verwijderd.', 'success');
                    self.expenditures.splice(self.expenditures.findIndex((e2) => e.getId() === e2.getId()), 1);
                    self.$emit('expenditure-deleted', e);
                    self.$forceUpdate();
                    self.loading = false;
                }).catch((reason) => {
                    (<any>$).notify('Kon uitgave niet verwijderen: ' + reason.message, 'warn');
                    self.loading = false;
                });
        },

        formatDate(d) {
            if(this.format === 'calendar') {
                return d.calendar();
            } else if(this.format === 'fromNow') {
                return d.fromNow();
            }

            return d.format(this.format);
        },
    },
    computed: {
        capped() {
            if(this.limit < 0) {
                return this.expenditures;
            }

            return this.expenditures.slice(0, this.limit);
        },
    },
    mounted: function() {
    },
    watch: {
    },
};
</script>

