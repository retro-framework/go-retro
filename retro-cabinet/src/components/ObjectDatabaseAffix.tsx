import './ObjectDatabaseAffix.css';

import * as React from 'react';

interface IObjectDatabaseAffixProps {
    refHash: string;
}

/* tslint:disable:no-console */

class ObjectDatabaseAffix extends React.Component<IObjectDatabaseAffixProps, any> {
    constructor(props: any) {
        super(props);
        this.state = {
            affix: {
                foo: "bar",
                hash: this.props.refHash,
            },
        };
    }
    public componentWillReceiveProps() {
        console.log("got here")
        if (this.state.affix.hash !== "") {
            fetch(`/obj/${this.state.affix.hash}`)
            // .then((response) => {
            //     console.log(response);
            // })
        }
    }
    public render() {
        return (
            <div className="obdv__affix">
                <span className="odbv__affixEntityName">{this.state.affix.hash}</span>
                <span className="odbv__affixEventHash">evhash</span>
            </div>
        )
    }
}

export default ObjectDatabaseAffix;
