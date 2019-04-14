module Main exposing (main)

import Browser exposing (element)
import Html exposing (..)
import Html.Attributes exposing (..)
import Html.Events exposing (..)
import Http
import Json.Decode exposing (Decoder, bool, field, list, map2, map4, maybe, oneOf, string)
import Json.Encode as Encode
import String exposing (join, split, toLower)


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


type alias Model =
    { docs : List Doc
    , current : Doc
    }


type Msg
    = Noop
    | FetchDocs (Result Http.Error (List Doc))
    | FetchDoc (Result Http.Error Doc)
    | SetTitle String
    | SetContent String
    | SelectDoc String
    | SubmitDoc
    | HandleResult (Result Http.Error ())


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
        Noop ->
            ( model, Cmd.none )

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
        [ navBar
        , main_ []
            [ aside [] [ showDocs model.docs ]
            , docForm model.current
            , previewDoc model.current
            ]
        ]


docForm : Doc -> Html Msg
docForm doc =
    Html.form [ class "edit" ]
        [ fieldset []
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
        , button [ type_ "button", onClick SubmitDoc ] [ text "Create Doc" ]
        ]


navBar : Html Msg
navBar =
    nav []
        [ h2 [] [ text "Doc Shelf" ] ]


previewDoc : Doc -> Html Msg
previewDoc doc =
    article []
        [ h1 [] [ text <| "Preview: " ++ doc.title ]
        , p [] [ text doc.content ]
        ]


showDocs : List Doc -> Html Msg
showDocs docs =
    ul [] (List.map (\d -> li [] [ a [ href "#", onClick (SelectDoc d.path) ] [ text d.title ] ]) docs)


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


main =
    element
        { init = init
        , update = update
        , subscriptions = subscriptions
        , view = view
        }
