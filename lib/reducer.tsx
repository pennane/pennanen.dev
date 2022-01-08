export interface AppState {
    finished: boolean
}

export enum AppActionType {
    ANIMATION_COMPLETE
}

export interface AppAction {
    type: AppActionType
    payload?: any
}

export const initialState: AppState = {
    finished: false
}

export const reducer = (state: AppState, action: AppAction): AppState => {
    console.log('reducer')

    switch (action.type) {
        case AppActionType.ANIMATION_COMPLETE:
            return {
                ...state,
                finished: true
            }

        default:
            return state
    }
}
