import { Record } from 'pocketbase';
import { Subscription, BehaviorSubject } from 'rxjs';
import pb from './PocketBase';


const stores: {
    [collection: string]: any
} = {}

function NewStore<T>(collection: string, transformer: (r: Record) => T): void {
    if (stores[collection]) {
        return;
    }
    const realTimeStore: {
        subject: BehaviorSubject<{
            [key: string]: T
        } | undefined>,
        init: () => Promise<void>,
        subscribe: (cb: (r: {
            [key: string]: T
        } | undefined) => void) => Subscription
    } = {
        subject: new BehaviorSubject<{
            [key: string]: T
        } | undefined>(undefined),
        subscribe: () => { return new Subscription() },
        init: () => { return Promise.reject() }
    };

    realTimeStore.init = () => {
        realTimeStore.subject = new BehaviorSubject<{
            [key: string]: T
        } | undefined>(undefined);
        return pb.collection(collection).getFullList()
            .then((data) => {
                var d: {
                    [key: string]: T
                } = {};
                for (let index = 0; index < data.length; index++) {
                    const element = data[index];
                    d[element.id] = transformer(element);
                }

                realTimeStore.subject.next(d);
                pb.collection(collection).subscribe("*", (data) => {

                    console.log(data);

                    var val = realTimeStore.subject.getValue();
                    if (!val) {
                        return;
                    }
                    switch (data.action) {
                        case "create":
                            val[data.record.id] = transformer(data.record);
                            break;

                        case "update":
                            val[data.record.id] = transformer(data.record);
                            break;

                        case "delete":
                            delete val[data.record.id];
                            break;

                        default:
                            break;
                    }

                    realTimeStore.subject.next(val);
                });
            });
    }

    realTimeStore.subscribe = (cb: (r: {
        [key: string]: T
    } | undefined) => void) => {
        return realTimeStore.subject.subscribe({
            next: cb
        })
    }

    realTimeStore.init();

    stores[collection] = realTimeStore;
}

function GetStore<T>(collection: string): Promise<RealTimeStore<T>> {
    var p = new Promise<RealTimeStore<T>>((resolve, reject) => {
        var checkStore = function () {
            var store = stores[collection];

            if (!store) {
                setTimeout(() => {
                    checkStore();
                }, 1000);
                return;
            }

            resolve(store);
        }
        checkStore();
    });

    return p;
}

interface RealTimeStore<T> {
    subscribe: (cb: (r: {
        [key: string]: T
    } | undefined) => void) => Subscription
}

const def = {
    NewStore: NewStore,
    GetStore: GetStore
};

export default def;