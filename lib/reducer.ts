export type AppState = {
  animations: {
    [animationKey: string]: boolean
  }
}

enum AppActionType {
  ANIMATION_COMPLETE,
}

type Action<TType extends AppActionType, TPayload> = {
  type: TType
  payload: TPayload
}

type AnimationCompletedAction = Action<AppActionType.ANIMATION_COMPLETE, string>

export type AppAction = AnimationCompletedAction

export const initialState: AppState = {
  animations: {},
}

export const reducer = (state: AppState, action: AppAction): AppState => {
  switch (action.type) {
    case AppActionType.ANIMATION_COMPLETE: {
      return {
        ...state,
        animations: { ...state.animations, [action.payload]: true },
      }
    }

    default:
      return state
  }
}

export const animationComplete = (key: string): AppAction => {
  return {
    type: AppActionType.ANIMATION_COMPLETE,
    payload: key,
  }
}
