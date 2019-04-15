module Main exposing (main)

import Browser exposing (element)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode exposing (Decoder, bool, field, list, map2, map4, maybe, oneOf, string)
import Json.Encode as Encode
import Markdown
import String exposing (join, split, toLower)


main =
    element
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }


type alias Doc =
    { path : String
    , title : String
    , isDir : Bool
    , content : String
    }


type alias SuccessResponse =
    { message : Maybe String
    , error : Maybe String
    }


type MenuStatus
    = Opened
    | Closed
    | Unused


type alias Model =
    { docs : List Doc
    , current : Doc
    , menuStatus : MenuStatus
    }


type Msg
    = ResetDoc
    | FetchDocs (Result Http.Error (List Doc))
    | FetchDoc (Result Http.Error Doc)
    | SetTitle String
    | SetContent String
    | SelectDoc String
    | SubmitDoc
    | HandleResult (Result Http.Error ())
    | ToggleMenu


newDoc : Doc
newDoc =
    { path = ""
    , title = ""
    , isDir = False
    , content = ""
    }


init : () -> ( Model, Cmd Msg )
init _ =
    ( { docs = []
      , current = newDoc
      , menuStatus = Unused
      }
    , fetchDocs
    )


fetchDocs : Cmd Msg
fetchDocs =
    Http.get
        { url = "http://localhost:1337/api/doc/list"
        , expect = Http.expectJson FetchDocs docsDecoder
        }


getDoc : String -> Cmd Msg
getDoc path =
    Http.get
        { url = "http://localhost:1337/api/doc/" ++ path
        , expect = Http.expectJson FetchDoc docDecoder
        }


subscriptions : Model -> Sub Msg
subscriptions model =
    Sub.none


update : Msg -> Model -> ( Model, Cmd Msg )
update msg model =
    case msg of
        ResetDoc ->
            ( { model | current = newDoc }, Cmd.none )

        FetchDocs result ->
            case result of
                Ok docs ->
                    ( { model | docs = docs }, Cmd.none )

                Err _ ->
                    ( model, Cmd.none )

        FetchDoc result ->
            case result of
                Ok doc ->
                    ( { model | current = doc }, Cmd.none )

                Err _ ->
                    ( model, Cmd.none )

        SetTitle title ->
            let
                updated =
                    model.current
                        |> setTitle title
            in
            ( { model | current = updated }, Cmd.none )

        SetContent content ->
            let
                updated =
                    model.current
                        |> setContent content
            in
            ( { model | current = updated }, Cmd.none )

        SelectDoc path ->
            let
                selected =
                    findDoc model.docs path
            in
            ( { model | current = selected }, getDoc path )

        SubmitDoc ->
            ( model
            , Http.post
                { url = "http://localhost:1337/api/doc"
                , body = Http.jsonBody <| encodeDoc model.current
                , expect = Http.expectWhatever HandleResult
                }
            )

        HandleResult result ->
            case result of
                Ok res ->
                    ( model, fetchDocs )

                Err _ ->
                    ( model, Cmd.none )

        ToggleMenu ->
            let
                newStatus =
                    case model.menuStatus of
                        Opened ->
                            Closed

                        Closed ->
                            Opened

                        Unused ->
                            Opened
            in
            ( { model | menuStatus = newStatus }, Cmd.none )


findDoc : List Doc -> String -> Doc
findDoc docs path =
    case List.head <| List.filter (\d -> d.path == path) docs of
        Just doc ->
            doc

        Nothing ->
            newDoc


setTitle : String -> Doc -> Doc
setTitle title doc =
    let
        path =
            title
                |> toLower
                |> split " "
                |> join "-"
    in
    { doc | title = title, path = path }


setContent : String -> Doc -> Doc
setContent content doc =
    { doc | content = content }


view : Model -> Html Msg
view model =
    div []
        [ navBar model.menuStatus
        , menu model.docs model.menuStatus
        , main_ []
            [ docForm model.current
            , previewDoc model.current
            ]
        ]


docForm : Doc -> Html Msg
docForm doc =
    Html.form [ class "edit" ]
        [ h1 [] [ text "Edit" ]
        , fieldset []
            [ legend [] [ text "Title" ]
            , input [ onInput SetTitle, value doc.title ] []
            ]
        , fieldset []
            [ legend [] [ text "Path" ]
            , p [] [ text doc.path ]
            ]
        , fieldset []
            [ legend [] [ text "Content" ]
            , textarea [ onInput SetContent, value doc.content ] []
            ]
        , div [ class "button-list" ]
            [ button [ type_ "button", onClick SubmitDoc, class "button", class "primary-button" ] [ text "Create Doc" ]
            , button [ type_ "button", onClick ResetDoc, class "button" ] [ text "Reset Form" ]
            ]
        ]


navBar : MenuStatus -> Html Msg
navBar status =
    nav []
        [ hamburger status
        , h2 [] [ text "Doc Shelf" ]
        ]


hamburger : MenuStatus -> Html Msg
hamburger status =
    let
        isActive =
            case status of
                Opened ->
                    "is-active"

                Closed ->
                    ""

                Unused ->
                    ""
    in
    button [ onClick ToggleMenu, class "hamburger", class "hamburger--arrow", class "menu-button", class isActive ]
        [ span [ class "hamburger-box" ]
            [ span [ class "hamburger-inner" ] [] ]
        ]


previewDoc : Doc -> Html Msg
previewDoc doc =
    article []
        [ h1 [] [ text <| "Quick Preview: " ++ doc.title ]
        , Markdown.toHtml [] doc.content
        ]


menu : List Doc -> MenuStatus -> Html Msg
menu docs status =
    let
        statusClass =
            case status of
                Opened ->
                    "slide-in"

                Closed ->
                    "slide-away"

                Unused ->
                    ""
    in
    aside [ class statusClass ]
        [ ul [] (List.map (\d -> li [] [ a [ href "#", onClick (SelectDoc d.path) ] [ text d.title ] ]) docs) ]


optionalField : String -> Decoder a -> a -> Decoder a
optionalField name try default =
    oneOf [ field name try, Json.Decode.succeed default ]


docDecoder : Decoder Doc
docDecoder =
    map4 Doc
        (field "path" string)
        (field "title" string)
        (field "isDir" bool)
        (optionalField "content" string "")


docsDecoder : Decoder (List Doc)
docsDecoder =
    list docDecoder


successDecoder : Decoder SuccessResponse
successDecoder =
    map2 SuccessResponse
        (maybe (field "message" string))
        (maybe (field "error" string))


encodeDoc : Doc -> Encode.Value
encodeDoc doc =
    Encode.object
        [ ( "path", Encode.string doc.path )
        , ( "title", Encode.string doc.title )
        , ( "content", Encode.string doc.content )
        ]
