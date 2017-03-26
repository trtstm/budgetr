import Entity from './entity';

class Category extends Entity {
    constructor(data?: { id?: number; name?: string }) {
        super(data ? data.id : undefined);

        if (!data) {
            data = {};
        }

        if (data.name) {
            this.name = data.name;
        }
    }

    getName(): string {
        return this.name;
    }

    setName(name: string) {
        this.name = name;
    }

    private name: string = '';

}

export default Category;