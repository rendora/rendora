package browser

var page = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=0"/>
    <meta http-equiv="X-UA-Compatible" content="IE=edge"/>
    <link rel="shortcut icon" href="//a.links123.cn/common/imgs/favicon.ico?2019092101" type="image/x-icon"/>
    <title>{{.Title}}</title>
    <style>
        .ie-warning {
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            bottom: 0;
            z-index: 100;
            background: #fff;
            padding-top: 69px;
        }

        .ie-header {
            width: 100%;
            padding: 20px 0;
            border-bottom: 1px solid #dadada;
            margin-top: -69px;
        }

        .ie-header img {
            display: block;
            height: 28px;
            margin: 0 auto;
        }

        .ie-body {
            background: #fbfbfb;
            padding-top: 20px;
            height: 100%;
        }

        .ie-tips {
            width: 800px;
            height: 400px;
            background: #fff;
            margin: 0 auto;
            text-align: center;
            -ms-filter: progid:DXImageTransform.Microsoft.Shadow(color=#EEEEEE, direction=0, strength=6) progid:DXImageTransform.Microsoft.Shadow(color=#EEEEEE, direction=90, strength=6) progid:DXImageTransform.Microsoft.Shadow(color=#EEEEEE, direction=180, strength=6) progid:DXImageTransform.Microsoft.Shadow(color=#EEEEEE, direction=270, strength=6);
            filter: progid:DXImageTransform.Microsoft.Shadow(color=#EEEEEE, direction=0, strength=6) progid:DXImageTransform.Microsoft.Shadow(color=#EEEEEE, direction=90, strength=6) progid:DXImageTransform.Microsoft.Shadow(color=#EEEEEE direction=180, strength=6) progid:DXImageTransform.Microsoft.Shadow(color=#EEEEEE, direction=270, strength=6);
        }

        .ie-tips .divider {
            width: 680px;
            height: 1px;
            background: #ddd;
            margin: 0 auto;
        }

        .ie-tips .main-tip img {
            width: 49px;
            margin-top: 60px;
        }

        .ie-tips .tip {
            color: #1790ff;
            font-size: 16px;
            font-weight: 600;
            line-height: 26px;
        }

        .ie-tips .subtip {
            font-size: 16px;
            color: #1790ff;
            font-weight: 400;
            line-height: 26px;
            margin-bottom: 26px;
        }

        .ie-tips ul {
            margin-top: 48px;
            margin-left: 15px;
            padding-left: 0;
            list-style: none;
        }

        .ie-tips ul li {
            position: relative;
            float: left;
            margin-left: 32px;
            margin-right: 32px;
            width: 90px;
        }

        .ie-tips ul li a {
            text-decoration: none;
            border: none;
        }

        .ie-tips li .divider {
            width: 2px;
            height: 34px;
            background: #ddd;
            position: absolute;
            right: -32px;
            top: 20px;
        }

        .ie-tips li img {
            border: none;
            /* width: 40px; */
        }

        .ie-tips li p {
            font-size: 14px;
            color: #4a4a4a;
            line-height: 20px;
            margin-top: 4px;
            margin-bottom: 2px;
        }

        .ie-tips li .platform img {
            width: 16px;
        }
    </style>
</head>

<body>
<div class="ie-warning" id="idWarningDiv">
    <div class="ie-header">
        <img
                id="logo_img"
                src="{{.Logo}}"
                alt="logo"
        />
    </div>
    <div class="ie-body">
        <div class="ie-tips">
            <div class="main-tip">
                <img src="//a.links123.cn/common/imgs/ie/tips_icon@2x.png"/>
                <p
                        id="tip_text_one"
                        class="tip"
                        style="display: block;
              margin-block-start: 1em;
              margin-block-end: 1em;
              margin-inline-start: 0px;
              margin-inline-end: 0px;"
                >
                    {{.Suggest}}
                </p>
            </div>
            <div class="divider"></div>
            <ul>
                <li>
                    <a
                            href="{{.GoogleDownloadURL}}"
                            target="_blank"
                            rel="noopener noreferrer"
                    >
                        <img
                                src="//a.links123.cn/common/imgs/ie/logo_chrome@2x.png"
                                class="logo"
                        />
                        <p>Chrome</p>
                        <div class="platform">
                            <img src="//a.links123.cn/common/imgs/ie/icon_win@2x.png"/>
                            <img src="//a.links123.cn/common/imgs/ie/icon_apple@2x.png"/>
                            <img src="//a.links123.cn/common/imgs/ie/icon_lum@2x.png"/>
                        </div>
                    </a>
                    <div class="divider"></div>
                </li>
                <li>
                    <a
                            href="{{.FirefoxDownloadURL}}"
                            target="_blank"
                            rel="noopener noreferrer"
                    >
                        <img
                                src="//a.links123.cn/common/imgs/ie/logo_firefox@2x.png"
                                class="logo"
                        />
                        <p>Firefox</p>
                        <div class="platform">
                            <img src="//a.links123.cn/common/imgs/ie/icon_win@2x.png"/>
                            <img src="//a.links123.cn/common/imgs/ie/icon_apple@2x.png"/>
                            <img src="//a.links123.cn/common/imgs/ie/icon_lum@2x.png"/>
                        </div>
                    </a>
                    <div class="divider"></div>
                </li>
                <li>
                    <a
                            href="{{.EdgeDownloadURL}}"
                            target="_blank"
                            rel="noopener noreferrer"
                    >
                        <img
                                src="//a.links123.cn/common/imgs/ie/logo_edge@2x.png"
                                class="logo"
                        />
                        <p>Edge</p>
                        <div class="platform">
                            <img src="//a.links123.cn/common/imgs/ie/icon_win@2x.png"/>
                        </div>
                    </a>
                    <div class="divider"></div>
                </li>
                <li>
                    <a
                            href="{{.SafariDownloadURL}}"
                            target="_blank"
                            rel="noopener noreferrer"
                    >
                        <img
                                src="//a.links123.cn/common/imgs/ie/logo_safari@2x.png"
                                class="logo"
                        />
                        <p>Safari</p>
                        <div class="platform">
                            <img src="//a.links123.cn/common/imgs/ie/icon_apple@2x.png"/>
                        </div>
                    </a>
                    <div class="divider"></div>
                </li>
                <li>
                    <a
                            href="{{.OperaDownloadURL}}"
                            target="_blank"
                            rel="noopener noreferrer"
                    >
                        <img
                                src="//a.links123.cn/common/imgs/ie/logo_opera@2x.png"
                                class="logo"
                        />
                        <p>Opera</p>
                        <div class="platform">
                            <img src="//a.links123.cn/common/imgs/ie/icon_win@2x.png"/>
                            <img src="//a.links123.cn/common/imgs/ie/icon_apple@2x.png"/>
                            <img src="//a.links123.cn/common/imgs/ie/icon_lum@2x.png"/>
                        </div>
                    </a>
                </li>
            </ul>
        </div>
    </div>
</div>
</body>
</html>
`
