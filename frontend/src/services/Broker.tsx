import { BehaviorSubject, Subscription } from "rxjs";

interface IBroker {
    subjects: { [key: string]: BehaviorSubject<any>; },
    publish: (topic: string, data: any) => void,
    subscribe: (topic: string, cb: (value: any) => void, fac: () => {}) => Subscription,
}

const Broker: IBroker = {
    subjects: {},
    publish: () => { },
    subscribe: () => { return new Subscription() },
}

Broker.publish = (topic: string, data: any) => {
    if (!Broker.subjects[topic]) {
        Broker.subjects[topic] = new BehaviorSubject(undefined);
    }
    Broker.subjects[topic].next(data);
}

Broker.subscribe = (topic, cb, fac) => {
    if (!Broker.subjects[topic]) {
        Broker.subjects[topic] = new BehaviorSubject(undefined);
        if (fac) {
            fac();
        }
    }
    return Broker.subjects[topic].subscribe({
        next: cb
    });
}


export default Broker;