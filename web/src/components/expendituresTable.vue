<template>
    <table class="table">
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
                <td>{{categoryName(expenditure.getCategory())}}</td>
            </tr>
        </tbody>
    </table>
</template>

<script lang="ts">
import Vue from "vue";
import _ from 'lodash';
import moment from 'moment';

interface ExpendituresTable extends Vue {
    expenditures: Array<any>;
    categoryName(category: any): void;
}

export default {
    name: "expenditures-table",
    props: ['expenditures'],
    data () {
        return {
        };
    },
    methods: {
        categoryName (category: any) {
            if(category === null) {
                return '';
            }

            return category.getName();
        },
    },
    computed: {

    },
    mounted: function() {
    },
    watch: {
    },
} as Vue.ComponentOptions<ExpendituresTable>;
</script>

