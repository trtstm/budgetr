<template>
    <div>
    <expenditure-form></expenditure-form>

    <input id="daterange">

    <table>
        <thead>
            <tr>
                <th>Datum</th>
                <th>Uitgave</th>
                <th>Categorie</th>
            </tr>
        </thead>

        <tbody>
            <tr v-for="expenditure in expenditures">
                <td>{{expenditure.getDate().format('LL')}}</td>
                <td>{{expenditure.getAmount()}}</td>
                <td>{{expenditure.getCategory()}}</td>
            </tr>
        </tbody>
    </table>
    </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from 'lodash';
import moment from 'moment';
import * as $ from 'jquery';
import '@/static/vendor/bootstrap-daterangepicker/daterangepicker.js';
import '@/static/vendor/bootstrap-daterangepicker/daterangepicker.css';


import expenditureForm from '@/components/expenditureForm';

import Api from '@/api';
import Expenditure from '@/expenditure';
import Category from '@/category';


interface ExpendituresTable extends Vue {
    expenditures: Array<Expenditure>;
    start: moment.Moment;
    end: moment.Moment;
    fetch(): void;
    updateDateFields(): void;
}

export default {
    name: "expenditures-table",
    components: {expenditureForm},
    data () {
        return {
            start: moment().startOf('month'),
            end: moment().endOf('month'),
            expenditures: [],
        };
    },
    methods: {
        dateChanged: function() {
        },
        updateDateFields: function() {
        },
        fetch: function() {
            let self = this;
            Api.getExpenditures({start: self.start, end: self.end})
            .then((results) => {
                self.expenditures = results.data;
            });
        },
    },
    computed: {
    },
    mounted: function() {
        $('#daterange').daterangepicker({
        ranges: {
           'Today': [moment(), moment()],
           'Yesterday': [moment().subtract(1, 'days'), moment().subtract(1, 'days')],
           'Last 7 Days': [moment().subtract(6, 'days'), moment()],
           'Last 30 Days': [moment().subtract(29, 'days'), moment()],
           'This Month': [moment().startOf('month'), moment().endOf('month')],
           'Last Month': [moment().subtract(1, 'month').startOf('month'), moment().subtract(1, 'month').endOf('month')]
        }
        });
        this.fetch();
    },
    watch: {
        '$route': function() {
            //this.fetch();
        },
    },
} as Vue.ComponentOptions<ExpendituresTable>;
</script>

