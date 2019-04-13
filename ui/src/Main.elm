module Main exposing (main)

import Browser exposing (element)
import Html exposing (Html, h1, text)

type alias Doc =
    { path: String
    , title: String
    , isDir: Bool
    , content: String
    }

type alias Model =
    { docs: List Doc
    , current: Doc
    }

type Msg =
    Noop

init : () -> (Model, Cmd Msg)
init _ =
    (
        { docs = []
        , current =
            { path = ""
            , title = ""
            , isDir = False
            , content = ""
            }
        }, Cmd.none)

subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none

update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
    case msg of
        Noop -> (model, Cmd.none)

view : Model -> Html Msg
view model =
    h1 [] [ text "Hello, world!" ]

main = element
    { init = init
    , update = update
    , subscriptions = subscriptions
    , view = view
    }
