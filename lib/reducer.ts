export interface AppState {
    animations: {
        [animationKey: string]: boolean
    }
}

export enum AppActionType {
    ANIMATION_COMPLETE
}

export interface AppAction {
    type: AppActionType
    payload?: any
}

export const initialState: AppState = {
    animations: {}
}

export const reducer = (state: AppState, action: AppAction): AppState => {
    switch (action.type) {
        case AppActionType.ANIMATION_COMPLETE: {
            return { ...state, animations: { ...state.animations, [action.payload]: true } }
        }

        default:
            return state
    }
}

export const animationComplete = (key: string): AppAction => {
    return {
        type: AppActionType.ANIMATION_COMPLETE,
        payload: key
    }
}
