import axios from 'axios';
import moment from 'moment';

import Expenditure from './expenditure';
import Category from './category';

interface Results<T> {
    meta: {};
    data: Array<T>;
}

interface CategoryStat {
    id: number;
    name: string;
    total: number;
}

let root = '/api';

let endpoints = {
    expenditures: root + '/expenditures',
    categories: root + '/categories',
    categoryStats: root + '/stats/categories',
    generateExcel: root + '/exports/excel',
};

class Api {
    getExpenditures(args: { start?: moment.Moment, end?: moment.Moment, sort?: string, order?: string, limit?: number, offset?: number }): Promise<Results<Expenditure>> {
        let params: any = {};
        if (args.start) {
            params.start = args.start.format();
        }
        if (args.end) {
            params.end = args.end.format();
        }
        if (args.sort) {
            params.sort = args.sort;
        }
        if (args.limit) {
            params.limit = args.limit;
        }
        if (args.offset) {
            params.offset = args.offset;
        }
        if (args.order) {
            params.sort += '-' + args.order;
        }

        return this.logFailure('getExpenditures', axios.get(endpoints.expenditures, {
            params: params,
        }).then((response: any) => {
            response.data.data = response.data.data.map((raw: any) => {
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
        if (category !== null && category.getName() !== '') {
            params.category = category.getName();
        }

        return this.logFailure('createExpenditure', axios.post(endpoints.expenditures, params)
            .then((response: any) => {
                return this.transformExpenditure(response.data);
            }));
    }

    deleteExpenditure(expenditure: Expenditure): Promise<Expenditure> {
        return this.logFailure('deleteExpenditure', axios.delete(endpoints.expenditures + '/' + expenditure.getId().toString()));
    }

    createCategory(category: Category): Promise<Category> {
        return this.logFailure('createCategory', axios.post(endpoints.categories, {
            name: category.getName(),
        }).then((response: any) => {
            return this.transformCategory(response.data);
        }));
    }

    getCategories(): Promise<Results<Category>> {
        return this.logFailure('getCategories', axios.get(endpoints.categories, {
        }).then((response: any) => {
            response.data.data = response.data.data.map((raw: any) => {
                return this.transformCategory(raw);
            });

            return response.data;
        }));
    }

    getCategoryStats(args: { start: moment.Moment, end: moment.Moment }): Promise<Array<CategoryStat>> {
        return this.logFailure('getCategoryStats', axios.get(endpoints.categoryStats, {
            params: {
                start: (args.start.format()),
                end: (args.end.format()),
            }
        }).then((response: any) => {
            return response.data;
        }));
    }

    generateExcel(ranges: any): Promise<void> {
        let form = document.createElement('form');
        form.action = endpoints.generateExcel;
        form.method = 'post';
        form.style.display = 'none';

        let input = document.createElement('input')
        input.name = 'ranges';
        input.value = JSON.stringify(ranges);

        form.appendChild(input);

        document.body.appendChild(form);
        form.submit();

        form.remove();

        return Promise.resolve();
    }

    private logFailure(name: string, p: any): Promise<any> {
        let q = new Promise((resolve, reject) => {
            p.then((data) => {
                resolve(data);
            }).catch((reason: any) => {
                if (!reason || !reason.message) {
                    reason = { message: 'Something went wrong on the server.' };

                }
                console.log(name + ' failed: ' + reason.message);
                reject(reason);
            })
        });

        return q;
    }



    private transformCategory(raw: any): Category {
        return new Category(raw);
    }

    private transformExpenditure(raw: any): Expenditure {
        let expenditure = new Expenditure(raw);

        if (raw.category) {
            let category = this.transformCategory(raw.category);
            expenditure.setCategory(category);
        }

        return expenditure;
    }
}

export { endpoints };
export default new Api();