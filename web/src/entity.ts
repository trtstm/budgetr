abstract class Entity {
    constructor(id?: number) {
        if (id) {
            this.id = id;
        }
    }

    getId(): number {
        return this.id;
    }

    protected id: number = 0;
}

export default Entity;