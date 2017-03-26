import moment from 'moment';

import Entity from './entity';
import Category from './category';

class Expenditure extends Entity {
    constructor(data?: { id?: number; amount?: number, date?: string | moment.Moment }) {
        super(data ? data.id : undefined);

        if (!data) {
            data = {};
        }

        if (data.amount) {
            this.setAmount(data.amount);
        }

        if (data.date) {
            this.date = moment(data.date);
        }
    }

    getAmount(): number {
        return this.amount;
    }

    setAmount(amount: number) {
        this.amount = Number(amount);
    }

    getDate(): moment.Moment {
        return this.date;
    }

    setDate(date: string | moment.Moment) {
        this.date = moment(date);
    }

    setCategory(category: Category | null) {
        this.category = category;
    }

    getCategory(): Category | null {
        return this.category;
    }

    private amount: number = 0;
    private date: moment.Moment = moment();
    private category: Category | null = null;
}

export default Expenditure;