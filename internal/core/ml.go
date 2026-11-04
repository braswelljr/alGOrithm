package core

import (
	"fmt"
	"math"
	"strings"

	"github.com/braswelljr/alGOrithm/pkg/ml/classification"
	"github.com/braswelljr/alGOrithm/pkg/ml/clustering"
	"github.com/braswelljr/alGOrithm/pkg/ml/neural"
	"github.com/braswelljr/alGOrithm/pkg/ml/regression"
	"github.com/braswelljr/alGOrithm/pkg/trace"
)

// clusteringEntries is every entry of the clustering category.
func clusteringEntries() []Entry {
	return []Entry{
		dbscanEntry(),
		hierarchicalClusteringEntry(),
		kMeansEntry(),
		kMeansPlusPlusEntry(),
		silhouetteScoreEntry(),
	}
}

// regressionEntries is every entry of the regression category.
func regressionEntries() []Entry {
	return []Entry{
		gradientDescentEntry(),
		linearRegressionEntry(),
		logisticRegressionEntry(),
		multipleLinearRegressionEntry(),
		polynomialRegressionEntry(),
		regressionMetricsEntry(),
	}
}

// classificationEntries is every entry of the classification category.
func classificationEntries() []Entry {
	return []Entry{
		confusionMatrixEntry(),
		decisionTreeEntry(),
		gaussianNaiveBayesEntry(),
		kFoldCrossValidationEntry(),
		kNearestNeighborsEntry(),
		naiveBayesTextEntry(),
		perceptronEntry(),
		precisionRecallF1Entry(),
		trainTestSplitEntry(),
	}
}

// neuralEntries is every entry of the neural category.
func neuralEntries() []Entry {
	return []Entry{
		activationFunctionsEntry(),
		softmaxCrossEntropyEntry(),
		xorNetworkEntry(),
	}
}

const (
	blobsExample = "# x y: three groups\n1 1\n2 1\n1 2\n2 2\n3 2\n8 8\n9 8\n8 9\n9 9\n8 7\n15 2\n16 2\n15 3\n16 3\n17 3"
	lineExample  = "# x y, roughly y = 2x + 1\n0 1.1\n1 2.9\n2 5.2\n3 6.8\n4 9.1\n5 11\n6 12.8\n7 15.2"
	labeledPts   = "# x y class\n1 1 red\n2 1 red\n1 2 red\n2 3 red\n3 2 red\n7 7 blue\n8 7 blue\n7 8 blue\n8 9 blue\n9 8 blue\n4 8 green\n5 9 green\n4 9 green\n3 9 green"
	irisExample  = "# petal length, petal width (cm), species\n1.4 0.2 setosa\n1.3 0.2 setosa\n1.5 0.2 setosa\n1.7 0.4 setosa\n1.4 0.3 setosa\n4.7 1.4 versicolor\n4.5 1.5 versicolor\n4.9 1.5 versicolor\n4.0 1.3 versicolor\n4.6 1.5 versicolor\n6.0 2.5 virginica\n5.1 1.9 virginica\n5.9 2.1 virginica\n5.6 1.8 virginica\n5.8 2.2 virginica\n4.9 1.8 virginica"
	spamExample  = "spam: win a free prize now\nspam: free money click now\nspam: claim your prize money today\nspam: cheap offer win big\nham: lunch with the team at noon\nham: notes from the project meeting\nham: can we move the meeting to friday\nham: the report is attached for review"
)

// mlOpt reads the option o, its default when absent, and rejects a value
// outside o's bounds.
func mlOpt(in Input, o Option) (int, error) {
	v := in.opt(o.Name, o.Default)
	if v < o.Min || v > o.Max {
		return 0, inputError(fmt.Sprintf("%s must be from %d to %d", o.Label, o.Min, o.Max))
	}
	return v, nil
}

// mlOpts reads every option of e at once, by name.
func mlOpts(e *Entry, in Input) (map[string]int, error) {
	out := map[string]int{}
	for _, o := range e.Options {
		v, err := mlOpt(in, o)
		if err != nil {
			return nil, err
		}
		out[o.Name] = v
	}
	return out, nil
}

// pointsOf draws rows in the plane by their first two features, rounded;
// one feature is drawn on the x axis.
func pointsOf(rows [][]float64) []Point {
	out := make([]Point, len(rows))
	for i, r := range rows {
		out[i].X = int(math.Round(r[0]))
		if len(r) > 1 {
			out[i].Y = int(math.Round(r[1]))
		}
	}
	return out
}

// lossCurve samples a loss curve to at most 40 points, as thousandths for
// Values (drawn as bars) and as lines of text.
func lossCurve(loss []float64) ([]int, []string) {
	step := max(1, len(loss)/40)
	var vals []int
	var lines []string
	for i := 0; i < len(loss); i += step {
		vals = append(vals, int(math.Round(math.Min(loss[i], 1e6)*1000)))
		lines = append(lines, fmt.Sprintf("epoch %d: loss %.4f", i+1, loss[i]))
	}
	if (len(loss)-1)%step != 0 {
		last := len(loss) - 1
		vals = append(vals, int(math.Round(math.Min(loss[last], 1e6)*1000)))
		lines = append(lines, fmt.Sprintf("epoch %d: loss %.4f", last+1, loss[last]))
	}
	return vals, lines
}

func fmtVec(v []float64) string {
	parts := make([]string, len(v))
	for i, x := range v {
		parts[i] = fmt.Sprintf("%.3g", x)
	}
	return "(" + strings.Join(parts, ", ") + ")"
}

// parseSetup is snippet code that parses text with pkg.fn into vars.
func parseSetup(vars, pkg, fn, text string) string {
	return fmt.Sprintf("%s, err := %s.%s(%s)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", vars, pkg, fn, "`"+text+"`")
}

var clusteringK = Option{"k", "clusters", 3, 1, 50}

var clusteringEx = &Input{Text: blobsExample, Options: map[string]int{"k": 3, "seed": 1}}

func kMeansEntry() Entry {
	km := entry("k-means", "K-Means Clustering", "ml/clustering/kmeans.go", "KMeans")
	km.Summary, km.Time, km.Space, km.Input, km.Traced, km.Example = "Put each point with its nearest center, move each center to the mean of its points, repeat until nothing moves.", "O(n * k * d) per round", "O(n + k * d)", Points, true, clusteringEx
	km.Options = []Option{clusteringK, seedOption}
	km.Limits = fmt.Sprintf("at most %d points", clustering.MaxPoints)
	km.Run = func(in Input, r trace.Recorder) (Output, error) {
		o, err := mlOpts(&km, in)
		if err != nil {
			return Output{}, err
		}
		pts, err := clustering.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		res, err := clustering.TraceKMeans(pts, o["k"], int64(o["seed"]), r)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Points: pointsOf(pts), Values: res.Labels, Summary: map[string]float64{"inertia": res.Inertia, "rounds": float64(res.Iterations)}}
		for c, ctr := range res.Centroids {
			out.Lines = append(out.Lines, fmt.Sprintf("cluster %d: center %s", c, fmtVec(ctr)))
		}
		return out, nil
	}
	km.Call = func(in Input) (string, string) {
		return parseSetup("pts", "clustering", "Parse", in.Text), fmt.Sprintf("clustering.KMeans(pts, %d, %d)", in.opt("k", 3), in.opt("seed", 1))
	}
	return km
}

func kMeansPlusPlusEntry() Entry {
	pp := entry("k-means-plus-plus", "K-Means++ Seeding", "ml/clustering/kmeans.go", "KMeansPlusPlus")
	pp.Summary, pp.Time, pp.Space, pp.Input, pp.Example = "Pick starting centers far apart: each next one with probability proportional to its squared distance from the nearest chosen.", "O(n * k * d)", "O(n + k * d)", Points, clusteringEx
	pp.Options = []Option{clusteringK, seedOption}
	pp.Run = func(in Input, _ trace.Recorder) (Output, error) {
		o, err := mlOpts(&pp, in)
		if err != nil {
			return Output{}, err
		}
		pts, err := clustering.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		cs, err := clustering.KMeansPlusPlus(pts, o["k"], int64(o["seed"]))
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Points: pointsOf(pts)}
		for c, ctr := range cs {
			out.Lines = append(out.Lines, fmt.Sprintf("center %d: %s", c, fmtVec(ctr)))
		}
		return out, nil
	}
	pp.Call = func(in Input) (string, string) {
		return parseSetup("pts", "clustering", "Parse", in.Text), fmt.Sprintf("clustering.KMeansPlusPlus(pts, %d, %d)", in.opt("k", 3), in.opt("seed", 1))
	}
	return pp
}

func dbscanEntry() Entry {
	db := entry("dbscan", "DBSCAN", "ml/clustering/dbscan.go", "DBSCAN")
	db.Summary, db.Time, db.Space, db.Input, db.Traced = "Clusters by density: points with enough neighbors within a radius grow a cluster; the rest is noise. Finds the number of clusters itself.", "O(n^2 * d)", "O(n)", Points, true
	db.Example = &Input{Text: blobsExample + "\n12 12\n0 9", Options: map[string]int{"eps": 15, "min": 3}}
	db.Options = []Option{{"eps", "radius, in tenths", 15, 1, 100_000}, {"min", "neighbors for a core point", 3, 1, 1000}}
	db.Run = func(in Input, r trace.Recorder) (Output, error) {
		o, err := mlOpts(&db, in)
		if err != nil {
			return Output{}, err
		}
		pts, err := clustering.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		labels, err := clustering.TraceDBSCAN(pts, float64(o["eps"])/10, o["min"], r)
		if err != nil {
			return Output{}, wrap(err)
		}
		clusters, noise := 0, 0
		for _, l := range labels {
			clusters = max(clusters, l+1)
			if l == clustering.Noise {
				noise++
			}
		}
		return Output{Points: pointsOf(pts), Values: labels, Number: ptr(clusters), Summary: map[string]float64{"clusters": float64(clusters), "noise points": float64(noise)}}, nil
	}
	db.Call = func(in Input) (string, string) {
		return parseSetup("pts", "clustering", "Parse", in.Text), fmt.Sprintf("clustering.DBSCAN(pts, %g, %d)", float64(in.opt("eps", 15))/10, in.opt("min", 3))
	}
	return db
}

func hierarchicalClusteringEntry() Entry {
	hc := entry("hierarchical-clustering", "Agglomerative Hierarchical Clustering", "ml/clustering/hierarchical.go", "Agglomerative")
	hc.Summary, hc.Time, hc.Space, hc.Input, hc.Traced = "Start with every point alone and keep joining the two nearest clusters; the merges form a dendrogram you can cut into k clusters.", "O(n^3)", "O(n^2)", Points, true
	hc.Example = &Input{Text: blobsExample, Options: map[string]int{"linkage": 0, "k": 3}}
	hc.Options = []Option{{"linkage", "linkage: 0 single, 1 complete, 2 average", 0, 0, 2}, clusteringK}
	hc.Limits = "at most 300 points"
	hc.Run = func(in Input, r trace.Recorder) (Output, error) {
		o, err := mlOpts(&hc, in)
		if err != nil {
			return Output{}, err
		}
		pts, err := clustering.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		if len(pts) > 300 {
			return Output{}, inputError("at most 300 points")
		}
		merges, err := clustering.TraceAgglomerative(pts, clustering.Linkage(o["linkage"]), r)
		if err != nil {
			return Output{}, wrap(err)
		}
		labels, err := clustering.Cut(merges, len(pts), min(o["k"], len(pts)))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Points: pointsOf(pts), Values: labels, Lines: clustering.Dendrogram(merges, len(pts))}, nil
	}
	hc.Call = func(in Input) (string, string) {
		return parseSetup("pts", "clustering", "Parse", in.Text), fmt.Sprintf("clustering.Agglomerative(pts, clustering.Linkage(%d))", in.opt("linkage", 0))
	}
	return hc
}

func silhouetteScoreEntry() Entry {
	sil := entry("silhouette-score", "Silhouette Score", "ml/clustering/silhouette.go", "Silhouette")
	sil.Summary, sil.Time, sil.Space, sil.Input, sil.Example = "How well each point sits in its cluster, from -1 to 1, here for k-means at every k from 2 to the option: the best k scores highest.", "O(n^2 * d)", "O(n)", Points, &Input{Text: blobsExample, Options: map[string]int{"k": 6, "seed": 1}}
	sil.Options = []Option{{"k", "largest k to try", 6, 2, 20}, seedOption}
	sil.Run = func(in Input, _ trace.Recorder) (Output, error) {
		o, err := mlOpts(&sil, in)
		if err != nil {
			return Output{}, err
		}
		pts, err := clustering.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Points: pointsOf(pts), Summary: map[string]float64{}}
		best, bestK := -2.0, 0
		for kk := 2; kk <= min(o["k"], len(pts)-1); kk++ {
			res, err := clustering.KMeans(pts, kk, int64(o["seed"]))
			if err != nil {
				return Output{}, wrap(err)
			}
			s, _, err := clustering.Silhouette(pts, res.Labels)
			if err != nil {
				continue
			}
			out.Lines = append(out.Lines, fmt.Sprintf("k = %d: silhouette %.3f", kk, s))
			if s > best {
				best, bestK = s, kk
			}
		}
		if bestK == 0 {
			return Output{}, inputError("give at least three points")
		}
		out.Number = ptr(bestK)
		out.Summary["best k"], out.Summary["best silhouette"] = float64(bestK), best
		return out, nil
	}
	sil.Call = func(in Input) (string, string) {
		setup := parseSetup("pts", "clustering", "Parse", in.Text) +
			fmt.Sprintf("res, err := clustering.KMeans(pts, 3, %d)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", in.opt("seed", 1))
		return setup, "clustering.Silhouette(pts, res.Labels)"
	}
	return sil
}

var regressionEx = &Input{Text: lineExample}

func regressionXy(in Input) ([]float64, []float64, error) {
	xs, ys, err := regression.Parse(in.Text)
	if err != nil {
		return nil, nil, wrap(err)
	}
	return regression.Column(xs, 0), ys, nil
}

func regressionSetupXY(in Input) string { return parseSetup("xs, ys", "regression", "Parse", in.Text) }

func regressionFitSummary(x, y []float64, predict func(float64) float64) map[string]float64 {
	preds := make([]float64, len(x))
	for i, v := range x {
		preds[i] = predict(v)
	}
	mse, _ := regression.MSE(y, preds)
	r2, _ := regression.R2(y, preds)
	return map[string]float64{"mean squared error": mse, "R^2": r2}
}

func regressionWithPoints(x, y []float64) []Point {
	rows := make([][]float64, len(x))
	for i := range x {
		rows[i] = []float64{x[i], y[i]}
	}
	return pointsOf(rows)
}

func linearRegressionEntry() Entry {
	lin := entry("linear-regression", "Linear Regression", "ml/regression/linear.go", "Linear")
	lin.Summary, lin.Time, lin.Space, lin.Input, lin.Example = "The straight line through the data with the smallest squared errors, in closed form: slope = cov(x, y) / var(x).", "O(n)", "O(1)", Points, regressionEx
	lin.Run = func(in Input, _ trace.Recorder) (Output, error) {
		x, y, err := regressionXy(in)
		if err != nil {
			return Output{}, err
		}
		s, b, err := regression.Linear(x, y)
		if err != nil {
			return Output{}, wrap(err)
		}
		sum := regressionFitSummary(x, y, func(v float64) float64 { return s*v + b })
		sum["slope"], sum["intercept"] = s, b
		return Output{Points: regressionWithPoints(x, y), Lines: []string{fmt.Sprintf("y = %.4g x + %.4g", s, b)}, Summary: sum}, nil
	}
	lin.Call = func(in Input) (string, string) {
		return regressionSetupXY(in), "regression.Linear(regression.Column(xs, 0), ys)"
	}
	return lin
}

func multipleLinearRegressionEntry() Entry {
	mul := entry("multiple-linear-regression", "Multiple Linear Regression", "ml/regression/linear.go", "Multiple")
	mul.Summary, mul.Time, mul.Space, mul.Input = "A plane (or hyperplane) through data with several features, by solving the normal equations (X^T X) c = X^T y.", "O(n * d^2 + d^3)", "O(d^2)", Text
	mul.Example = &Input{Text: "# size (m^2), rooms, age (years), price (thousands)\n50 2 30 150\n70 3 20 210\n90 3 10 270\n60 2 25 175\n120 4 5 360\n80 3 15 240\n100 4 12 295\n65 2 40 170"}
	mul.Run = func(in Input, _ trace.Recorder) (Output, error) {
		xs, ys, err := regression.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		c, err := regression.Multiple(xs, ys)
		if err != nil {
			return Output{}, wrap(err)
		}
		preds := make([]float64, len(xs))
		for i, row := range xs {
			preds[i] = c[0]
			for j, v := range row {
				preds[i] += c[j+1] * v
			}
		}
		terms := []string{fmt.Sprintf("%.4g", c[0])}
		for j := 1; j < len(c); j++ {
			terms = append(terms, fmt.Sprintf("%.4g x%d", c[j], j))
		}
		mse, _ := regression.MSE(ys, preds)
		r2, _ := regression.R2(ys, preds)
		return Output{Lines: []string{"y = " + strings.Join(terms, " + ")}, Summary: map[string]float64{"mean squared error": mse, "R^2": r2}}, nil
	}
	mul.Call = func(in Input) (string, string) { return regressionSetupXY(in), "regression.Multiple(xs, ys)" }
	return mul
}

func gradientDescentEntry() Entry {
	gd := entry("gradient-descent", "Gradient Descent (linear regression)", "ml/regression/gradient.go", "GradientDescent")
	gd.Summary, gd.Time, gd.Space, gd.Input, gd.Example = "Learn a line by stepping downhill on the error, epoch after epoch; the loss curve shows learning, or diverging when the rate is too high.", "O(n * epochs)", "O(epochs)", Points, &Input{Text: lineExample, Options: map[string]int{"rate": 20, "epochs": 300}}
	gd.Options = []Option{{"rate", "learning rate, in thousandths", 20, 1, 10_000}, {"epochs", "epochs", 300, 1, 100_000}}
	gd.Run = func(in Input, _ trace.Recorder) (Output, error) {
		o, err := mlOpts(&gd, in)
		if err != nil {
			return Output{}, err
		}
		x, y, err := regressionXy(in)
		if err != nil {
			return Output{}, err
		}
		f, err := regression.GradientDescent(x, y, float64(o["rate"])/1000, o["epochs"])
		if err != nil {
			return Output{}, wrap(err)
		}
		vals, lines := lossCurve(f.Loss)
		lines = append([]string{fmt.Sprintf("learned y = %.4g x + %.4g", f.Slope, f.Intercept)}, lines...)
		return Output{Points: regressionWithPoints(x, y), Values: vals, Lines: lines, Summary: map[string]float64{"slope": f.Slope, "intercept": f.Intercept, "final loss": f.Loss[len(f.Loss)-1]}}, nil
	}
	gd.Call = func(in Input) (string, string) {
		return regressionSetupXY(in), fmt.Sprintf("regression.GradientDescent(regression.Column(xs, 0), ys, %g, %d)", float64(in.opt("rate", 20))/1000, in.opt("epochs", 300))
	}
	return gd
}

func polynomialRegressionEntry() Entry {
	poly := entry("polynomial-regression", "Polynomial Regression", "ml/regression/linear.go", "Polynomial")
	poly.Summary, poly.Time, poly.Space, poly.Input = "A curve c0 + c1 x + c2 x^2 + ... by least squares on the powers of x; raise the degree to see overfitting.", "O(n * degree^2 + degree^3)", "O(n * degree)", Points
	poly.Example = &Input{Text: "# x y, a parabola with noise\n-3 9.2\n-2 4.1\n-1 0.8\n0 0.1\n1 1.2\n2 3.9\n3 9.1\n4 15.8", Options: map[string]int{"degree": 2}}
	poly.Options = []Option{{"degree", "degree", 2, 1, 10}}
	poly.Run = func(in Input, _ trace.Recorder) (Output, error) {
		o, err := mlOpts(&poly, in)
		if err != nil {
			return Output{}, err
		}
		x, y, err := regressionXy(in)
		if err != nil {
			return Output{}, err
		}
		c, err := regression.Polynomial(x, y, o["degree"])
		if err != nil {
			return Output{}, wrap(err)
		}
		var terms []string
		for j, v := range c {
			terms = append(terms, fmt.Sprintf("%.4g x^%d", v, j))
		}
		return Output{Points: regressionWithPoints(x, y), Lines: []string{"y = " + strings.Join(terms, " + ")}, Summary: regressionFitSummary(x, y, func(v float64) float64 { return regression.Eval(c, v) })}, nil
	}
	poly.Call = func(in Input) (string, string) {
		return regressionSetupXY(in), fmt.Sprintf("regression.Polynomial(regression.Column(xs, 0), ys, %d)", in.opt("degree", 2))
	}
	return poly
}

func logisticRegressionEntry() Entry {
	lg := entry("logistic-regression", "Logistic Regression", "ml/regression/logistic.go", "FitLogistic")
	lg.Summary, lg.Time, lg.Space, lg.Input = "A yes-or-no model: the sigmoid of a weighted sum gives P(yes), learned by gradient descent on the log loss.", "O(n * d * epochs)", "O(d + epochs)", Text
	lg.Example = &Input{Text: "# hours studied, hours slept, passed (0 or 1)\n1 4 0\n2 5 0\n3 4 0\n2 8 0\n4 6 0\n5 7 1\n6 5 1\n7 8 1\n8 6 1\n6 8 1\n3 7 0\n5 4 0", Options: map[string]int{"rate": 300, "epochs": 2000}}
	lg.Options = []Option{{"rate", "learning rate, in thousandths", 300, 1, 10_000}, {"epochs", "epochs", 2000, 1, 100_000}}
	lg.Run = func(in Input, _ trace.Recorder) (Output, error) {
		o, err := mlOpts(&lg, in)
		if err != nil {
			return Output{}, err
		}
		xs, ys, err := regression.Parse(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		labels, err := regression.Binary(ys)
		if err != nil {
			return Output{}, wrap(err)
		}
		m, err := regression.FitLogistic(xs, labels, float64(o["rate"])/1000, o["epochs"])
		if err != nil {
			return Output{}, wrap(err)
		}
		right := 0
		for i, row := range xs {
			if m.Predict(row) == labels[i] {
				right++
			}
		}
		vals, lines := lossCurve(m.Loss)
		lines = append([]string{"weights (bias first): " + fmtVec(m.Weights)}, lines...)
		return Output{Points: pointsOf(xs), Values: vals, Lines: lines, Summary: map[string]float64{"training accuracy": float64(right) / float64(len(xs)), "final log loss": m.Loss[len(m.Loss)-1]}}, nil
	}
	lg.Call = func(in Input) (string, string) {
		setup := regressionSetupXY(in) + "labels, err := regression.Binary(ys)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t"
		return setup, fmt.Sprintf("regression.FitLogistic(xs, labels, %g, %d)", float64(in.opt("rate", 300))/1000, in.opt("epochs", 2000))
	}
	return lg
}

func regressionMetricsEntry() Entry {
	met := entry("regression-metrics", "Mean Squared Error and R^2", "ml/regression/metrics.go", "R2")
	met.Summary, met.Time, met.Space, met.Input = "How good predictions are: the mean squared error, and R^2, the share of the variance they explain.", "O(n)", "O(1)", Points
	met.Example = &Input{Text: "# actual predicted\n3 2.5\n5 5.2\n7 6.8\n9 9.5\n11 10.6"}
	met.Run = func(in Input, _ trace.Recorder) (Output, error) {
		actual, pred, err := regressionXy(in)
		if err != nil {
			return Output{}, err
		}
		mse, err := regression.MSE(actual, pred)
		if err != nil {
			return Output{}, wrap(err)
		}
		r2, _ := regression.R2(actual, pred)
		return Output{Points: regressionWithPoints(actual, pred), Summary: map[string]float64{"mean squared error": mse, "root mean squared error": math.Sqrt(mse), "R^2": r2}}, nil
	}
	met.Call = func(in Input) (string, string) {
		return regressionSetupXY(in), "regression.R2(regression.Column(xs, 0), ys)"
	}
	return met
}

func classificationLabeled(in Input) ([][]float64, []string, error) {
	xs, ls, err := classification.Parse(in.Text)
	return xs, ls, wrap(err)
}

func classificationSetupLabeled(in Input) string {
	return parseSetup("xs, labels", "classification", "Parse", in.Text)
}

func classificationSetupQuery(in Input) string {
	return classificationSetupLabeled(in) + fmt.Sprintf("query, err := classification.ParseRow(%q)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", in.Other)
}

func classificationConfusion(in Input) (classification.Confusion, error) {
	c, err := classification.ConfusionMatrix(strings.Fields(in.Text), strings.Fields(in.Other))
	return c, wrap(err)
}

var classificationConfEx = &Input{Text: "cat cat cat dog dog dog dog bird bird cat", Other: "cat dog cat dog dog cat dog bird cat cat"}

func kNearestNeighborsEntry() Entry {
	knn := entry("k-nearest-neighbors", "K-Nearest Neighbors", "ml/classification/knn.go", "KNN")
	knn.Summary, knn.Time, knn.Space, knn.Input, knn.Traced = "Classify a new point by a vote of the k training points nearest it. No training at all: every query does the work.", "O(n * d + n log n) per query", "O(n)", TwoTexts, true
	knn.Example = &Input{Text: labeledPts, Other: "5 6", Options: map[string]int{"k": 3}}
	knn.Options = []Option{{"k", "neighbors", 3, 1, 100}}
	knn.Run = func(in Input, r trace.Recorder) (Output, error) {
		o, err := mlOpts(&knn, in)
		if err != nil {
			return Output{}, err
		}
		xs, ls, err := classificationLabeled(in)
		if err != nil {
			return Output{}, err
		}
		q, err := classification.ParseRow(in.Other)
		if err != nil {
			return Output{}, wrap(err)
		}
		label, near, err := classification.TraceKNN(xs, ls, q, o["k"], r)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Points: pointsOf(xs), Text: label, Values: near}
		for _, i := range near {
			out.Lines = append(out.Lines, fmt.Sprintf("neighbor %d %s: %s", i, fmtVec(xs[i]), ls[i]))
		}
		return out, nil
	}
	knn.Call = func(in Input) (string, string) {
		return classificationSetupQuery(in), fmt.Sprintf("classification.KNN(xs, labels, query, %d)", in.opt("k", 3))
	}
	return knn
}

func gaussianNaiveBayesEntry() Entry {
	gnb := entry("gaussian-naive-bayes", "Gaussian Naive Bayes", "ml/classification/bayes.go", "FitGaussianNB")
	gnb.Summary, gnb.Time, gnb.Space, gnb.Input = "Per class, a bell curve for each feature; the most probable class wins, pretending the features are independent.", "O(n * d) to fit, O(classes * d) to predict", "O(classes * d)", TwoTexts
	gnb.Example = &Input{Text: irisExample, Other: "4.8 1.6"}
	gnb.Run = func(in Input, _ trace.Recorder) (Output, error) {
		xs, ls, err := classificationLabeled(in)
		if err != nil {
			return Output{}, err
		}
		q, err := classification.ParseRow(in.Other)
		if err != nil {
			return Output{}, wrap(err)
		}
		if len(q) != len(xs[0]) {
			return Output{}, inputError(fmt.Sprintf("the query needs %d features", len(xs[0])))
		}
		m, err := classification.FitGaussianNB(xs, ls)
		if err != nil {
			return Output{}, wrap(err)
		}
		label, scores := m.Predict(q)
		out := Output{Points: pointsOf(xs), Text: label}
		for c, name := range m.Classes {
			out.Lines = append(out.Lines, fmt.Sprintf("%s: prior %.2f, mean %s, log score %.2f", name, m.Prior[c], fmtVec(m.Mean[c]), scores[c]))
		}
		return out, nil
	}
	gnb.Call = func(in Input) (string, string) {
		return classificationSetupQuery(in) + "m, err := classification.FitGaussianNB(xs, labels)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", "m.Predict(query)"
	}
	return gnb
}

func naiveBayesTextEntry() Entry {
	tnb := entry("naive-bayes-text", "Naive Bayes Spam Filter", "ml/classification/bayes.go", "FitTextNB")
	tnb.Summary, tnb.Time, tnb.Space, tnb.Input = "Count words per class, then score a new message by how likely each class is to produce its words (with add-one smoothing).", "O(words) to fit and to predict", "O(classes * vocabulary)", TwoTexts
	tnb.Example = &Input{Text: spamExample, Other: "win free money now"}
	tnb.Run = func(in Input, _ trace.Recorder) (Output, error) {
		docs, ls, err := classification.ParseDocs(in.Text)
		if err != nil {
			return Output{}, wrap(err)
		}
		m, err := classification.FitTextNB(docs, ls)
		if err != nil {
			return Output{}, wrap(err)
		}
		label, scores := m.Predict(in.Other)
		out := Output{Text: label}
		for c, name := range m.Classes {
			out.Lines = append(out.Lines, fmt.Sprintf("%s: log score %.3f", name, scores[c]))
		}
		return out, nil
	}
	tnb.Call = func(in Input) (string, string) {
		setup := parseSetup("docs, labels", "classification", "ParseDocs", in.Text) + "m, err := classification.FitTextNB(docs, labels)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t"
		return setup, fmt.Sprintf("m.Predict(%q)", in.Other)
	}
	return tnb
}

func perceptronEntry() Entry {
	per := entry("perceptron", "Perceptron", "ml/classification/perceptron.go", "FitPerceptron")
	per.Summary, per.Time, per.Space, per.Input = "The first learning machine: nudge a line toward every example it gets wrong. It stops if a line separates the classes, and never does on XOR.", "O(n * d * epochs)", "O(d + epochs)", Text
	per.Example = &Input{Text: "# x y class: yes above the line y = x\n1 3 yes\n2 5 yes\n3 6 yes\n4 7 yes\n2 1 no\n4 2 no\n5 3 no\n6 4 no", Options: map[string]int{"epochs": 100}}
	per.Options = []Option{{"epochs", "epochs", 100, 1, 10_000}}
	per.Run = func(in Input, _ trace.Recorder) (Output, error) {
		o, err := mlOpts(&per, in)
		if err != nil {
			return Output{}, err
		}
		xs, ls, err := classificationLabeled(in)
		if err != nil {
			return Output{}, err
		}
		classes := classification.Classes(ls)
		if len(classes) != 2 {
			return Output{}, inputError("a perceptron separates exactly two classes")
		}
		p, err := classification.FitPerceptron(xs, classification.Signs(ls, classes[1]), o["epochs"])
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Points: pointsOf(xs), Values: p.Mistakes, Lines: []string{fmt.Sprintf("+1 is %q; weights (bias first): %s", classes[1], fmtVec(p.Weights))}}
		for e, m := range p.Mistakes {
			out.Lines = append(out.Lines, fmt.Sprintf("epoch %d: %d mistakes", e+1, m))
		}
		out.Bool = ptr(p.Mistakes[len(p.Mistakes)-1] == 0)
		return out, nil
	}
	per.Call = func(in Input) (string, string) {
		_, ls, _ := classification.Parse(in.Text)
		pos := ""
		if cs := classification.Classes(ls); len(cs) > 1 {
			pos = cs[1]
		}
		return classificationSetupLabeled(in), fmt.Sprintf("classification.FitPerceptron(xs, classification.Signs(labels, %q), %d)", pos, in.opt("epochs", 100))
	}
	return per
}

func decisionTreeEntry() Entry {
	tree := entry("decision-tree", "Decision Tree (CART)", "ml/classification/tree.go", "BuildTree")
	tree.Summary, tree.Time, tree.Space, tree.Input = "Ask the yes-or-no question that best splits the classes (lowest Gini impurity), then ask again on each side.", "O(d * n^2 * depth)", "O(n) per level", Text
	tree.Example = &Input{Text: irisExample, Options: map[string]int{"depth": 3}}
	tree.Options = []Option{{"depth", "maximum depth", 3, 0, 30}}
	tree.Run = func(in Input, _ trace.Recorder) (Output, error) {
		o, err := mlOpts(&tree, in)
		if err != nil {
			return Output{}, err
		}
		xs, ls, err := classificationLabeled(in)
		if err != nil {
			return Output{}, err
		}
		n, err := classification.BuildTree(xs, ls, o["depth"])
		if err != nil {
			return Output{}, wrap(err)
		}
		right := 0
		for i, row := range xs {
			if n.Predict(row) == ls[i] {
				right++
			}
		}
		return Output{Points: pointsOf(xs), Lines: n.Lines(), Summary: map[string]float64{"training accuracy": float64(right) / float64(len(xs))}}, nil
	}
	tree.Call = func(in Input) (string, string) {
		return classificationSetupLabeled(in), fmt.Sprintf("classification.BuildTree(xs, labels, %d)", in.opt("depth", 3))
	}
	return tree
}

func confusionMatrixEntry() Entry {
	cm := entry("confusion-matrix", "Confusion Matrix", "ml/classification/evaluate.go", "ConfusionMatrix")
	cm.Summary, cm.Time, cm.Space, cm.Input, cm.Example = "A table of the true class against the predicted one: the diagonal is right, everything else shows which classes get mixed up.", "O(n + classes^2)", "O(classes^2)", TwoTexts, classificationConfEx
	cm.Run = func(in Input, _ trace.Recorder) (Output, error) {
		c, err := classificationConfusion(in)
		if err != nil {
			return Output{}, err
		}
		t := &TableOutput{Rows: len(c.Classes), Cols: len(c.Classes)}
		for i, row := range c.M {
			t.Cells = append(t.Cells, row...)
			t.RowLabels = append(t.RowLabels, "true "+c.Classes[i])
			t.ColLabels = append(t.ColLabels, "said "+c.Classes[i])
		}
		_, acc := c.Scores()
		return Output{Table: t, Summary: map[string]float64{"accuracy": acc}}, nil
	}
	cm.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("classification.ConfusionMatrix(%#v, %#v)", strings.Fields(in.Text), strings.Fields(in.Other))
	}
	return cm
}

func precisionRecallF1Entry() Entry {
	prf := entry("precision-recall-f1", "Precision, Recall and F1", "ml/classification/evaluate.go", "Confusion.Scores")
	prf.Summary, prf.Time, prf.Space, prf.Input, prf.Example = "Per class: of the predictions, how many were right (precision); of the truth, how much was found (recall); and their balance (F1).", "O(classes^2)", "O(classes)", TwoTexts, classificationConfEx
	prf.Run = func(in Input, _ trace.Recorder) (Output, error) {
		c, err := classificationConfusion(in)
		if err != nil {
			return Output{}, err
		}
		scores, acc := c.Scores()
		out := Output{Summary: map[string]float64{"accuracy": acc}}
		for _, s := range scores {
			out.Lines = append(out.Lines, fmt.Sprintf("%s: precision %.2f, recall %.2f, F1 %.2f (%d examples)", s.Class, s.Precision, s.Recall, s.F1, s.Support))
		}
		return out, nil
	}
	prf.Call = func(in Input) (string, string) {
		return fmt.Sprintf("c, err := classification.ConfusionMatrix(%#v, %#v)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\t", strings.Fields(in.Text), strings.Fields(in.Other)), "c.Scores()"
	}
	return prf
}

func trainTestSplitEntry() Entry {
	split := entry("train-test-split", "Train/Test Split", "ml/classification/evaluate.go", "TrainTestSplit")
	split.Summary, split.Time, split.Space, split.Input = "Shuffle the examples and hold some back, so a model is judged on data it never trained on.", "O(n)", "O(n)", Simulation
	split.Example = &Input{Options: map[string]int{"n": 10, "test": 30, "seed": 1}}
	split.Options = []Option{{"n", "examples", 10, 2, 100_000}, {"test", "test share, percent", 30, 1, 99}, seedOption}
	split.Run = func(in Input, _ trace.Recorder) (Output, error) {
		o, err := mlOpts(&split, in)
		if err != nil {
			return Output{}, err
		}
		train, test, err := classification.TrainTestSplit(o["n"], float64(o["test"])/100, int64(o["seed"]))
		if err != nil {
			return Output{}, wrap(err)
		}
		return Output{Values: test, Lines: []string{fmt.Sprintf("train (%d): %v", len(train), train), fmt.Sprintf("test (%d): %v", len(test), test)}}, nil
	}
	split.Call = func(in Input) (string, string) {
		return "", fmt.Sprintf("classification.TrainTestSplit(%d, %g, %d)", in.opt("n", 10), float64(in.opt("test", 30))/100, in.opt("seed", 1))
	}
	return split
}

func kFoldCrossValidationEntry() Entry {
	cv := entry("k-fold-cross-validation", "K-Fold Cross-Validation", "ml/classification/evaluate.go", "CrossValidateKNN")
	cv.Summary, cv.Time, cv.Space, cv.Input = "Split the data into k folds, test on each in turn while training on the rest, and average: every example is tested exactly once.", "O(n^2 * d) for k-NN", "O(n)", Text
	cv.Example = &Input{Text: irisExample, Options: map[string]int{"folds": 4, "k": 3, "seed": 1}}
	cv.Options = []Option{{"folds", "folds", 4, 2, 100}, {"k", "neighbors for k-NN", 3, 1, 100}, seedOption}
	cv.Run = func(in Input, _ trace.Recorder) (Output, error) {
		o, err := mlOpts(&cv, in)
		if err != nil {
			return Output{}, err
		}
		xs, ls, err := classificationLabeled(in)
		if err != nil {
			return Output{}, err
		}
		acc, err := classification.CrossValidateKNN(xs, ls, o["folds"], o["k"], int64(o["seed"]))
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{}
		sum := 0.0
		for f, a := range acc {
			sum += a
			out.Lines = append(out.Lines, fmt.Sprintf("fold %d: accuracy %.2f", f+1, a))
		}
		out.Summary = map[string]float64{"mean accuracy": sum / float64(len(acc))}
		return out, nil
	}
	cv.Call = func(in Input) (string, string) {
		return classificationSetupLabeled(in), fmt.Sprintf("classification.CrossValidateKNN(xs, labels, %d, %d, %d)", in.opt("folds", 4), in.opt("k", 3), in.opt("seed", 1))
	}
	return cv
}

func xorNetworkEntry() Entry {
	xor := entry("xor-network", "Neural Network Learning XOR", "ml/neural/mlp.go", "MLP.Train")
	xor.Summary, xor.Time, xor.Space, xor.Input = "A two-layer network learns XOR by backpropagation: errors flow backward through the weights by the chain rule.", "O(epochs * n * weights)", "O(units + epochs)", Simulation
	xor.Example = &Input{Options: map[string]int{"hidden": 4, "epochs": 5000, "rate": 500, "seed": 1}}
	xor.Options = []Option{{"hidden", "hidden units", 4, 1, 32}, {"epochs", "epochs", 5000, 1, 50_000}, {"rate", "learning rate, in thousandths", 500, 1, 10_000}, seedOption}
	xor.Run = func(in Input, _ trace.Recorder) (Output, error) {
		o, err := mlOpts(&xor, in)
		if err != nil {
			return Output{}, err
		}
		m, err := neural.NewMLP([]int{2, o["hidden"], 1}, int64(o["seed"]))
		if err != nil {
			return Output{}, wrap(err)
		}
		xs, ys := neural.XOR()
		loss, err := m.Train(xs, ys, float64(o["rate"])/1000, o["epochs"])
		if err != nil {
			return Output{}, wrap(err)
		}
		vals, lines := lossCurve(loss)
		var preds []string
		right := 0
		for i, x := range xs {
			p, _ := m.Predict(x)
			preds = append(preds, fmt.Sprintf("%v XOR %v = %v, network says %.3f", x[0], x[1], ys[i][0], p[0]))
			if math.Round(p[0]) == ys[i][0] {
				right++
			}
		}
		return Output{Values: vals, Lines: append(preds, lines...), Bool: ptr(right == len(xs)), Summary: map[string]float64{"final loss": loss[len(loss)-1], "correct": float64(right)}}, nil
	}
	xor.Call = func(in Input) (string, string) {
		setup := fmt.Sprintf("m, err := neural.NewMLP([]int{2, %d, 1}, %d)\n\tif err != nil {\n\t\tpanic(err)\n\t}\n\txs, ys := neural.XOR()\n\t", in.opt("hidden", 4), in.opt("seed", 1))
		return setup, fmt.Sprintf("m.Train(xs, ys, %g, %d)", float64(in.opt("rate", 500))/1000, in.opt("epochs", 5000))
	}
	return xor
}

func activationFunctionsEntry() Entry {
	act := entry("activation-functions", "Activation Functions", "ml/neural/neural.go", "Activation")
	act.Summary, act.Time, act.Space, act.Input = "Sigmoid, tanh and ReLU, and their slopes, from x = -4 to 4: why ReLU keeps gradients alive and sigmoid lets them fade.", "O(1) each", "O(1)", Simulation
	act.Example = &Input{}
	act.Run = func(_ Input, _ trace.Recorder) (Output, error) {
		out := Output{Lines: []string{"x: sigmoid (slope), tanh (slope), relu (slope)"}}
		for x := -4.0; x <= 4; x++ {
			parts := make([]string, 0, len(neural.Activations))
			for _, a := range neural.Activations {
				parts = append(parts, fmt.Sprintf("%.3f (%.3f)", a.F(x), a.DF(x)))
			}
			out.Lines = append(out.Lines, fmt.Sprintf("%+.0f: %s", x, strings.Join(parts, ", ")))
		}
		out.Summary = map[string]float64{"largest sigmoid slope": neural.Sigmoid.DF(0), "largest tanh slope": neural.Tanh.DF(0)}
		return out, nil
	}
	act.Call = func(Input) (string, string) {
		return "", "neural.Sigmoid.F(1), neural.Tanh.F(1), neural.ReLU.F(1)"
	}
	return act
}

func softmaxCrossEntropyEntry() Entry {
	sm := entry("softmax-cross-entropy", "Softmax and Cross-Entropy", "ml/neural/neural.go", "Softmax")
	sm.Summary, sm.Time, sm.Space, sm.Input, sm.Params = "Turn scores into probabilities that sum to 1, and score them against the right answer: the loss a classifier learns from.", "O(n)", "O(n)", AnyInts, []string{"target"}
	sm.Example = &Input{Values: []int{2, 1, 0, 3}, Target: 3}
	sm.Run = func(in Input, _ trace.Recorder) (Output, error) {
		if err := values(in, "turn into probabilities", 1000); err != nil {
			return Output{}, err
		}
		x := make([]float64, len(in.Values))
		for i, v := range in.Values {
			x[i] = float64(v)
		}
		p, err := neural.Softmax(x)
		if err != nil {
			return Output{}, wrap(err)
		}
		ce, err := neural.CrossEntropy(p, in.Target)
		if err != nil {
			return Output{}, wrap(err)
		}
		out := Output{Summary: map[string]float64{"cross-entropy": ce}}
		for i, v := range p {
			out.Lines = append(out.Lines, fmt.Sprintf("class %d: score %d, probability %.4f", i, in.Values[i], v))
			out.Values = append(out.Values, int(math.Round(v*1000)))
		}
		return out, nil
	}
	sm.Call = func(in Input) (string, string) {
		f := make([]string, len(in.Values))
		for i, v := range in.Values {
			f[i] = fmt.Sprint(v)
		}
		return "", "neural.Softmax([]float64{" + strings.Join(f, ", ") + "})"
	}
	return sm
}
