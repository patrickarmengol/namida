# namida

Artsy rain and haikus in your terminal.

![demo.gif](demo.gif?raw=true)

## Usage

```
git clone https://github.com/patrickarmengol/namida.git
cd namida
go build
./namida -fps 60 -raindropRate 1 -linger 400 -cooldown 400
```

## Todo

-   accept user specified haiku files
-   allow multiple haikus to be up at once
-   clean up error handling
-   maybe put it up on the aur
-   colors???

## Notes

I got the idea from rewatching an episode of Nichijou while fixing my linux rice.

This was hastily/shoddily put together. The code looks bad. I might make it look not so bad sometime in the future.

## License

`namida` is distributed under the terms of any of the following licenses:

-   [Apache-2.0](https://spdx.org/licenses/Apache-2.0.html)
-   [MIT](https://spdx.org/licenses/MIT.html)

### Third-Party Licenses

This project uses the following third-party packages:

-   [`gdamore/tcell`](https://github.com/gdamore/tcell) (Apache-2.0 License): This package is used for drawing cells in the terminal. See [Apache-2.0 License](https://spdx.org/licenses/Apache-2.0.html) for details.
