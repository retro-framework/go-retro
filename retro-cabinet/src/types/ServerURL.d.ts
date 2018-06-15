import { IStateErrorable, IStateLoadable, ICheckpoint } from './index'

/*
 * Redux State
 */
export interface IState {
    readonly url: URL
}

/*
 * Component Props
 */
interface IPropsFromState {
    url: URL
}

// tslint:disable-next-line:no-empty-interface
interface IPropsFromDispatch {
    refreshRefsFromURL: (_: URL) => void
}

// tslint:disable-next-line:no-empty-interface
export type IProps = IPropsFromState | IPropsFromDispatch;