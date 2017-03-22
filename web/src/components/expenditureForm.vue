<template>
    <div>
        <input ref="date"
               type="date"
               v-on:change="dateChanged">
    
        <input type="text"
               v-model="amount">

        <input type="text"
               v-model="category">
    
        <input type="button"
               v-on:click="create">
    </div>
</template>

<script lang="ts">
import Vue from "vue";
import _ from 'lodash';
import moment from 'moment';

import api from '@/api';
import Expenditure from '@/expenditure';
import Category from '@/category';


interface ExpenditureForm extends Vue {
    amount: string;
    category: string;
    date: moment.Moment;
    create(): void;
    dateChanged(): void;
    updateDateField(): void;
}

export default {
    name: "expenditure-form",
    data () {
        return {
            date: moment().startOf('day'),
            amount: 0,
            category: '',
        };
    },
    methods: {
        dateChanged: function() {
            this.date = moment((<any>this).$refs.date.valueAsDate);
        },
        updateDateField: function() {
            (<any>this).$refs.date.valueAsDate = this.date.toDate();
        },
        create: function() {
            let expenditure = new Expenditure({amount: parseFloat(this.amount), date: this.date});
            if(this.category != '') {
                expenditure.setCategory(new Category({name: this.category}));
            }
            api.createExpenditure(expenditure);
        },
    },
    mounted () {
        this.updateDateField();
    },
    watch: {
        '$route': function() {
            //this.fetch();
        },
    },
} as Vue.ComponentOptions<ExpenditureForm>;
</script>