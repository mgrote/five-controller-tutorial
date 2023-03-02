#### demo
- make a fresh deploy of branch `controller-tutorial`
- `k patch outlet light-one --type='merge' -p '{"spec":{"switch":"ON"}}'`
- `k patch outlet light-three --type='merge' -p '{"spec":{"switch":"ON"}}'`
- `k patch location tutorial-room --type='merge' -p '{"spec":{"mood":"BRIGHT"}}'`
- `k patch location tutorial-room --type='merge' -p '{"spec":{"mood":"DARK"}}'`
- `k patch location tutorial-room --type='merge' -p '{"spec":{"mood":"DONTKNOW"}}'`

- `k patch location tutorial-room --type='merge' -p '{"spec":{"mood":"BRIGHT"}}'`
- `k delete location tutorial-room`

- `k patch outlet light-one --type='merge' -p '{"spec":{"switch":"ON"}}'`
- `k patch outlet light-three --type='merge' -p '{"spec":{"switch":"ON"}}'`

- search for pid with `ps -aux | grep '/manager'`
- connect with delve `dlv attach 124536 --listen=:2345 --headless --api-version=2`