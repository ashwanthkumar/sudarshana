# devmerge_2k18

VSCode plugin that works with the `sudarsana-web` is available on https://github.com/ashwanthkumar/vscode-go/tree/0.6.93-autobot.

### Datasets
[Stackoverflow dump](https://data.stackexchange.com/stackoverflow/query/913091/dump-question-answers-for-a-tag) of all [`go`](https://stackoverflow.com/questions/tagged/go) tagged questions and all their answers.

You can either download the CSV and have a mult-line CSV parser parse that output. It's a multi-line because of question / answer body. Or you can run the Query directly and download the query revision page (like [this](https://data.stackexchange.com/stackoverflow/revision/913091/1136390/dump-question-answers-for-a-tag) page, warning the page is 100M in size). The page contains the resultSet as a JSON embedded in it. Assume you've the JSON extracted from the HTML Page, you can generate the dataset using the following command

```
$ wget "https://data.stackexchange.com/stackoverflow/revision/913091/1136390/dump-question-answers-for-a-tag" -O dump-question-answers-for-a-tag-with-title
$ ... extract the JSON content alone in the file
$ jq .resultSets[].rows[] dump-question-answers-for-a-tag-with-title | jq -c '{"title": .[0], "qId": .[1], "aId": .[2], "tags": .[3], "qScore": .[4], "aScore": .[5], "question": .[6], "answer": .[7], "views": .[8], "nrOfCommentsInAnswer": .[9], "qDate": .[10], "aDate": .[11]}' > go-so-questions-and-all-answers-with-title.json
```
