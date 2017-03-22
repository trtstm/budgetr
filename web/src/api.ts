import axios from 'axios';
import _ from 'lodash';
import moment from 'moment';

import Expenditure from './expenditure';
import Category from './category';

interface Results<T> {
    meta: {};
    data: Array<T>;
}

let root = '/api';

let endpoints = {
    expenditures: root + '/expenditures',
    categories: root + '/categories',
};

class Api {
    getExpenditures(args: { start: moment.Moment, end: moment.Moment }): Promise<Results<Expenditure>> {
        return this.logFailure('getExpenditures', axios.get(endpoints.expenditures, {
            params: {
                start: (args.start.format()),
                end: (args.end.format()),
            }
        }).then((response: any) => {
            response.data.data = _.map(response.data.data, (raw: any) => {
                return this.transformExpenditure(raw);
            });

            return response.data;
        }));
    }

    createExpenditure(expenditure: Expenditure): Promise<Expenditure> {
        let params: any = {
            date: expenditure.getDate().format(),
            amount: expenditure.getAmount(),
        };

        let category = expenditure.getCategory();
        if(category !== null && category.getName() !== '') {
            params.category = category.getName();
        }

        return this.logFailure('createExpenditure', axios.post(endpoints.expenditures,params)
        .then((response: any) => {
            return this.transformExpenditure(response.data);
        }));
    }

    createCategory(category: Category): Promise<Category> {
        return this.logFailure('createCategory', axios.post(endpoints.categories, {
            name: category.getName(),
        }).then((response: any) => {
            return this.transformCategory(response.data);
        }));
    }

    private logFailure(name: string, p: any): Promise<any> {
        return p.catch((reason: any) => {
            if (!reason || !reason.message) {
                reason = { message: 'Something went wrong on the server.' };

            }
            console.log(name + ' failed: ' + reason.message);
            return reason;
        })
    }



    private transformCategory(raw: any): Category {
        return new Category(raw);
    }

    private transformExpenditure(raw: any): Expenditure {
        let expenditure = new Expenditure(raw);

        if(raw.category) {
            let category = this.transformCategory(raw.category);
            expenditure.setCategory(category);
        }

        return expenditure;
    }
}

export { endpoints };
export default new Api();